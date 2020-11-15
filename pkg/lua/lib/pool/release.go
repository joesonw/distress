package pool

import (
	"fmt"
	"io"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"
)

type Resource interface {
	Name() string
	Release() error
}

type ResourceFunc struct {
	name    string
	release func() error
}

func NewReleaseFunc(name string, f func() error) Resource {
	return &ResourceFunc{
		name:    name,
		release: f,
	}
}

func (f *ResourceFunc) Name() string {
	return f.name
}

func (f *ResourceFunc) Release() error {
	return f.release()
}

type ioReadCloserResource struct {
	io.ReadCloser
	name string
}

func (r *ioReadCloserResource) Name() string {
	return r.name
}

func (r *ioReadCloserResource) Release() error {
	return r.Close()
}

func NewIOReadCloserResource(name string, rc io.ReadCloser) Resource {
	return &ioReadCloserResource{
		ReadCloser: rc,
		name:       name,
	}
}

type ioWriteCloserResource struct {
	io.WriteCloser
	name string
}

func (r *ioWriteCloserResource) Name() string {
	return r.name
}

func (r *ioWriteCloserResource) Release() error {
	return r.Close()
}

func NewIOWriteCloserResource(name string, rc io.WriteCloser) Resource {
	return &ioWriteCloserResource{
		WriteCloser: rc,
		name:        name,
	}
}

type ioReadWriteCloserResource struct {
	io.ReadWriteCloser
	name string
}

func (r *ioReadWriteCloserResource) Name() string {
	return r.name
}

func (r *ioReadWriteCloserResource) Release() error {
	return r.Close()
}

func NewIOReadWriteCloserResource(name string, rc io.ReadWriteCloser) Resource {
	return &ioReadWriteCloserResource{
		ReadWriteCloser: rc,
		name:            name,
	}
}

type OSFile interface {
	Name() string
	Close() error
}

type osFileResource struct {
	OSFile
}

func (r *osFileResource) Name() string {
	return r.OSFile.Name()
}

func (r *osFileResource) Release() error {
	return r.Close()
}

func NewOSFileResource(f OSFile) Resource {
	return &osFileResource{f}
}

type Guard struct {
	id              int64
	pool            *ReleasePool
	resource        Resource
	suppressWarning bool
}

func (g *Guard) ID() int64 {
	return g.id
}

func (g *Guard) Name() string {
	return g.resource.Name()
}

func (g *Guard) Done() {
	atomic.AddInt64(&g.pool.statsCurrent, -1)
	g.pool.pool.Delete(g.id)
}

func (g *Guard) SuppressWarning(suppressWarning bool) *Guard {
	g.suppressWarning = suppressWarning
	return g
}

type ReleasePool struct {
	pool      *sync.Map
	idCounter int64
	logger    *zap.Logger

	statsTotal   int64
	statsCurrent int64
}

func NewRelease(logger *zap.Logger) *ReleasePool {
	return &ReleasePool{
		pool:   &sync.Map{},
		logger: logger.With(zap.String("lua-module", "ResourceReleasePool")),
	}
}

func (p *ReleasePool) Watch(r Resource) *Guard {
	id := atomic.AddInt64(&p.idCounter, 1)
	atomic.AddInt64(&p.statsCurrent, 1)
	atomic.AddInt64(&p.statsTotal, 1)
	g := &Guard{
		id:       id,
		pool:     p,
		resource: r,
	}
	p.pool.Store(id, g)
	return g
}

func (p *ReleasePool) Clean() {
	p.statsCurrent = 0
	p.pool.Range(func(key, value interface{}) bool {
		p.pool.Delete(key)
		g := value.(*Guard)
		err := g.resource.Release()
		if !g.suppressWarning {
			p.logger.Warn(fmt.Sprintf("resource#%d \"%s\" leaked, now releasing", g.id, g.resource.Name()))
		}
		if err != nil {
			p.logger.With(zap.Error(err)).Error(fmt.Sprintf("resource#%d \"%s\" leaked, release encountered error", g.id, g.resource.Name()))
		}
		return true
	})
}

func (p *ReleasePool) Total() int64 {
	return p.statsTotal
}

func (p *ReleasePool) Len() int64 {
	return p.statsCurrent
}
