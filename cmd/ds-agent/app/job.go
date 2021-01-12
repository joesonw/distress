package app

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"go.uber.org/zap"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libpool "github.com/joesonw/lte/pkg/lua/lib/pool"
	luavm "github.com/joesonw/lte/pkg/lua/vm"
	"github.com/joesonw/lte/pkg/stat"
)

type Job struct {
	logger       *zap.Logger
	fs           afero.Fs
	vms          []*luavm.VM
	concurrency  int
	global       *luacontext.Global
	statReporter stat.Reporter

	finishedAmount int64
	startedAt      time.Time
	totalAmount    int64
	totalDuration  time.Duration
}

func NewJob(
	logger *zap.Logger,
	fs afero.Fs,
	entry string,
	concurrency int,
	envs map[string]string,
	newFS func() afero.Fs,
	reporter stat.Reporter,
) (*Job, error) {
	source, err := afero.ReadFile(fs, entry)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open entry file")
	}

	proto, err := luavm.Compile(string(source), entry)
	if err != nil {
		return nil, errors.Wrap(err, "unable to compile")
	}

	global := luacontext.NewGlobal(reporter)
	vms := make([]*luavm.VM, concurrency)
	asyncPool := libpool.NewAsync(logger, 4, time.Second*30, 64)
	for i := 0; i < concurrency; i++ {

		vms[i] = luavm.New(logger, asyncPool, global, luavm.Parameters{
			EnvVars:    envs,
			Filesystem: afero.NewCopyOnWriteFs(fs, newFS()),
		})
		if err := vms[i].Load(proto); err != nil {
			return nil, err
		}
	}

	return &Job{
		logger:       logger,
		fs:           fs,
		vms:          vms,
		concurrency:  concurrency,
		statReporter: reporter,
		global:       global,
		startedAt:    time.Now(),
	}, nil
}

func (j *Job) RunDuration(duration time.Duration) {
	j.logger.Info(fmt.Sprintf("run in time constraint mode: %s, conncurency: %d", duration.String(), j.concurrency))
	j.totalDuration = duration
	stopAt := time.Now().Add(duration)
	j.Run(func() bool {
		return time.Now().After(stopAt)
	})
}

func (j *Job) RunAmount(amount int64) {
	j.logger.Info(fmt.Sprintf("run in amount mode: %d, conncurency: %d", amount, j.concurrency))
	var count int64
	j.totalAmount = amount
	j.Run(func() bool {
		return atomic.AddInt64(&count, 1) > amount
	})
}

func (j *Job) Run(shouldStop func() bool) {
	wg := &sync.WaitGroup{}
	var counter int64

	for _, vm := range j.vms {
		wg.Add(1)
		go func(vm *luavm.VM) {
			defer wg.Done()
			for {
				if shouldStop() {
					return
				}
				start := time.Now()
				if err := vm.Run(atomic.AddInt64(&counter, 1)); err != nil {
					j.logger.Error("error running script", zap.Error(err))
				}
				since := time.Since(start)
				j.statReporter.Report(stat.New("run").Field("cost", float64(since.Nanoseconds())))
				atomic.AddInt64(&j.finishedAmount, 1)
				vm.Reset()
			}
		}(vm)
	}

	wg.Wait()
}

func (j *Job) Close() {
}
