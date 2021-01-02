package async

import (
	"context"

	lua "github.com/yuin/gopher-lua"

	"github.com/joesonw/lte/pkg/lua/lib/pool"
)

func Deferred(L *lua.LState, asyncPool *pool.AsyncPool, f func(ctx context.Context) error) int {
	ch := make(chan error)
	asyncPool.Add(pool.AsyncTaskFunc(func(ctx context.Context) error {
		ch <- f(ctx)
		return nil
	}))

	L.Push(L.NewFunction(func(L *lua.LState) int {
		err := <-ch
		if err != nil {
			L.Push(lua.LString(err.Error()))
			return 1
		}
		return 0
	}))
	return 1
}

func DeferredResult(L *lua.LState, asyncPool *pool.AsyncPool, f func(ctx context.Context) (lua.LGFunction, error)) int {
	ch := make(chan struct{})
	var lf lua.LGFunction
	var err error
	asyncPool.Add(pool.AsyncTaskFunc(func(ctx context.Context) error {
		lf, err = f(ctx)
		ch <- struct{}{}
		return nil
	}))

	L.Push(L.NewFunction(func(L *lua.LState) int {
		<-ch
		if err != nil {
			L.Push(lua.LString(err.Error()))
		} else {
			L.Push(lua.LNil)
		}

		if lf == nil {
			return 1
		}

		n := lf(L)
		return n + 1
	}))
	return 1
}
