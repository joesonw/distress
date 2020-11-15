package app

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/montanaflynn/stats"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/uber-go/tally"
	"go.uber.org/zap"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	luavm "github.com/joesonw/distress/pkg/lua/vm"
)

type Job struct {
	logger      *zap.Logger
	fs          afero.Fs
	vms         []*luavm.VM
	concurrency int

	runData      stats.Float64Data
	runHistogram tally.Histogram
}

func newJob(
	logger *zap.Logger,
	fs afero.Fs,
	entry string,
	concurrency int,
	envs map[string]string,
	newFS func() afero.Fs,
	scope tally.Scope,
) (*Job, error) {
	source, err := afero.ReadFile(fs, entry)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open entry file")
	}

	proto, err := luavm.Compile(string(source))
	if err != nil {
		return nil, errors.Wrap(err, "unable to compile")
	}

	global := luacontext.NewGlobal()
	vms := make([]*luavm.VM, concurrency)
	for i := 0; i < concurrency; i++ {
		vms[i] = luavm.New(logger, global, luavm.Parameters{
			AsyncPoolConcurrency: 4,
			AsyncPoolTimeout:     time.Second * 30,
			AsyncPoolBufferSize:  64,
			EnvVars:              envs,
			Filesystem:           afero.NewCopyOnWriteFs(fs, newFS()),
		})
		if err := vms[i].Load(proto); err != nil {
			return nil, err
		}
	}

	runHistogram := scope.Histogram("run", tally.MustMakeExponentialDurationBuckets(time.Microsecond, 10, 10))

	return &Job{
		logger:       logger,
		fs:           fs,
		vms:          vms,
		concurrency:  concurrency,
		runHistogram: runHistogram,
	}, nil
}

func (j *Job) RunInfinity(ch chan os.Signal) {
	j.logger.Info(fmt.Sprintf("run in infinity mode, conncurency: %d", j.concurrency))
	done := false
	go func() {
		<-ch
		done = true
	}()
	j.run(func() bool {
		return done
	})
}

func (j *Job) RunDuration(duration time.Duration) {
	j.logger.Info(fmt.Sprintf("run in time constraint mode: %s, conncurency: %d", duration.String(), j.concurrency))
	stopAt := time.Now().Add(duration)
	j.run(func() bool {
		return time.Now().After(stopAt)
	})
}

func (j *Job) RunAmount(amount int64) {
	j.logger.Info(fmt.Sprintf("run in amount mode: %d, conncurency: %d", amount, j.concurrency))
	var count int64
	j.run(func() bool {
		return atomic.AddInt64(&count, 1) > amount
	})
}

func (j *Job) run(shouldStop func() bool) {
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
					j.logger.With(zap.Error(err)).Error("error running script")
				}
				since := time.Since(start)
				us := float64(since.Milliseconds())
				j.runData = append(j.runData, us)
				j.runHistogram.RecordDuration(since)
				vm.Reset()
			}
		}(vm)
	}

	wg.Wait()
}

func (j *Job) Report() {
}

func (j *Job) Close() {
}
