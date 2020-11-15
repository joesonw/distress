package pool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"

	goutil "github.com/joesonw/distress/pkg/util"
)

type AsyncTask interface {
	Do(context.Context) error
}

type AsyncTaskFunc func(context.Context) error

func (f AsyncTaskFunc) Do(ctx context.Context) error {
	return f(ctx)
}

type AsyncPool struct {
	mu           *sync.Mutex
	chTasks      chan AsyncTask
	chExit       []chan struct{}
	isRunning    bool
	concurrency  int
	timeout      time.Duration
	logger       *zap.Logger
	statsTotal   int64
	statsCurrent int64
	bufferSize   int
}

func NewAsync(logger *zap.Logger, concurrency int, timeout time.Duration, bufferSize int) *AsyncPool {
	return &AsyncPool{
		mu:          &sync.Mutex{},
		bufferSize:  bufferSize,
		concurrency: concurrency,
		timeout:     timeout,
		logger:      logger.With(zap.String("lua-module", "AsyncTaskPool")),
	}
}

func (p *AsyncPool) Add(task AsyncTask) {
	atomic.AddInt64(&p.statsCurrent, 1)
	atomic.AddInt64(&p.statsTotal, 1)
	p.chTasks <- task
}

func (p *AsyncPool) Start() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isRunning {
		return
	}
	p.chTasks = make(chan AsyncTask, p.bufferSize)
	p.isRunning = true
	p.chExit = make([]chan struct{}, p.concurrency)
	for i := 0; i < p.concurrency; i++ {
		ch := make(chan struct{})
		p.chExit[i] = ch
		go p.run(ch)
	}
}

func (p *AsyncPool) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.isRunning {
		return
	}
	p.isRunning = false
	for _, c := range p.chExit {
		c <- struct{}{}
	}
	p.chExit = nil
}

func (p *AsyncPool) run(chExit chan struct{}) {
	for {
		select {
		case <-chExit:
			return
		case task := <-p.chTasks:
			p.runTask(task)
		}
	}
}

func (p *AsyncPool) runTask(task AsyncTask) {
	ctx, cancel := goutil.NewOptionalTimeoutContext(p.timeout)
	defer cancel()
	defer atomic.AddInt64(&p.statsCurrent, -1)

	err := task.Do(ctx)
	if err != nil {
		p.logger.With(zap.Error(err)).Error("unable to handle async task")
	}
}

func (p *AsyncPool) Total() int64 {
	return p.statsTotal
}

func (p *AsyncPool) Len() int64 {
	return p.statsCurrent
}
