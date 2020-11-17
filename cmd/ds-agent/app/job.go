package app

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"go.uber.org/zap"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	luavm "github.com/joesonw/distress/pkg/lua/vm"
	"github.com/joesonw/distress/pkg/metrics"
)

type Job struct {
	logger      *zap.Logger
	fs          afero.Fs
	vms         []*luavm.VM
	concurrency int
	global      *luacontext.Global
	runMetric   metrics.Metric

	finishedAmount int64
	startedAt      time.Time
	totalAmount    int64
	totalDuration  time.Duration
}

func newJob(
	logger *zap.Logger,
	fs afero.Fs,
	entry string,
	concurrency int,
	envs map[string]string,
	newFS func() afero.Fs,
	reporter metrics.Reporter,
) (*Job, error) {
	source, err := afero.ReadFile(fs, entry)
	if err != nil {
		return nil, errors.Wrap(err, "unable to open entry file")
	}

	proto, err := luavm.Compile(string(source))
	if err != nil {
		return nil, errors.Wrap(err, "unable to compile")
	}

	global := luacontext.NewGlobal(reporter)
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

	runMetric := metrics.Gauge("run_us", nil)
	reporter.Collect(runMetric)

	return &Job{
		logger:      logger,
		fs:          fs,
		vms:         vms,
		concurrency: concurrency,
		runMetric:   runMetric,
		global:      global,
		startedAt:   time.Now(),
	}, nil
}

func (j *Job) RunDuration(duration time.Duration) {
	j.logger.Info(fmt.Sprintf("run in time constraint mode: %s, conncurency: %d", duration.String(), j.concurrency))
	j.totalDuration = duration
	stopAt := time.Now().Add(duration)
	j.run(func() bool {
		return time.Now().After(stopAt)
	})
}

func (j *Job) RunAmount(amount int64) {
	j.logger.Info(fmt.Sprintf("run in amount mode: %d, conncurency: %d", amount, j.concurrency))
	var count int64
	j.totalAmount = amount
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
					j.logger.Error("error running script", zap.Error(err))
				}
				since := time.Since(start)
				j.runMetric.Add(float64(since.Microseconds()))
				atomic.AddInt64(&j.finishedAmount, 1)
				vm.Reset()
			}
		}(vm)
	}

	wg.Wait()
}

func (j *Job) Stats() (*Stats, error) {
	data := metrics.MemoryData{}
	metrics.Memory(&data).Collect(j.runMetric)
	return &Stats{
		StartedAt:          j.startedAt.Unix(),
		Concurrency:        int64(j.concurrency),
		TotalAmount:        j.totalAmount,
		FinishedAmount:     j.finishedAmount,
		DurationMicro:      time.Since(j.startedAt).Microseconds(),
		TotalDurationMicro: j.totalDuration.Microseconds(),
		AverageCostMicro:   data.Gauge["run_us"].Mean,
	}, nil
}

func (j *Job) Close() {
}
