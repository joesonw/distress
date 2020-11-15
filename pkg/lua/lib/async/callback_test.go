package async_test

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libasync "github.com/joesonw/distress/pkg/lua/lib/async"
	test_util "github.com/joesonw/distress/pkg/lua/test-util"
)

var testTable = []struct {
	name   string
	script string
	before test_util.Before
}{{
	name: "deferred",
	script: `
		local aa, bb = d(), d()
		local a, b = aa(), bb()
		if a == "1" and b == "1" then
			error("a==1")
		elseif a == "2" and b == "2" then
			error("a==2")
		end
	`,
	before: func(t *testing.T, L *lua.LState, luaCtx *luacontext.Context) {
		l := &sync.Mutex{}
		c := sync.NewCond(l)

		go func() {
			time.Sleep(time.Second)
			c.L.Lock()
			c.Broadcast()
			c.L.Unlock()
		}()

		var i int64
		L.SetGlobal("d", L.NewClosure(func(L *lua.LState) int {
			return libasync.Deferred(L, luaCtx.AsyncPool(), func(ctx context.Context) error {
				c.L.Lock()
				c.Wait()
				c.L.Unlock()
				return fmt.Errorf("%d", atomic.AddInt64(&i, 1))
			})
		}))
	},
}, {
	name: "deferred_result",
	script: `
		local aa, bb = d(), d()
		local _, a = aa()
		local _, b = bb()
		assert(a == 1, "a == 1")
		assert(b == 2, "a == 2")
	`,
	before: func(t *testing.T, L *lua.LState, luaCtx *luacontext.Context) {
		l := &sync.Mutex{}
		c := sync.NewCond(l)

		go func() {
			time.Sleep(time.Second)
			c.L.Lock()
			c.Broadcast()
			c.L.Unlock()
		}()

		var i int64
		L.SetGlobal("d", L.NewClosure(func(L *lua.LState) int {
			return libasync.DeferredResult(L, luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
				c.L.Lock()
				c.Wait()
				c.L.Unlock()
				return func(L *lua.LState) int {
					L.Push(lua.LNumber(atomic.AddInt64(&i, 1)))
					return 1
				}, nil
			})
		}))
	},
}}

func Test(t *testing.T) {
	tests := make([]test_util.Testable, len(testTable))
	for i := range testTable {
		test := testTable[i]
		tests[i] = func(_ *testing.T) *test_util.Test {
			return test_util.New(test.name, test.script).
				Before(test.before)
		}
	}
	test_util.Run(t, tests...)
}
