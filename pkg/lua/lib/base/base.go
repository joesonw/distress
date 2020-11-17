package base

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/afero"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libasync "github.com/joesonw/distress/pkg/lua/lib/async"
	luautil "github.com/joesonw/distress/pkg/lua/util"
	"github.com/joesonw/distress/pkg/metrics"
)

type baseContext struct {
	luaCtx *luacontext.Context
	fs     afero.Fs
}

func Open(L *lua.LState, luaCtx *luacontext.Context, fs afero.Fs) {
	ud := L.NewUserData()
	ud.Value = &baseContext{
		luaCtx: luaCtx,
		fs:     fs,
	}
	for k, f := range funcs {
		L.SetGlobal(k, L.NewClosure(f, ud))
	}
}

func upContext(L *lua.LState) *baseContext {
	ud := L.CheckUserData(lua.UpvalueIndex(1))
	return ud.Value.(*baseContext)
}

var funcs = map[string]lua.LGFunction{
	"print":  lPrint,
	"group":  lGroup,
	"sleep":  lSleep,
	"import": lImport,
}

func lPrint(L *lua.LState) int {
	ctx := upContext(L)
	top := L.GetTop()
	s := ""
	for i := 1; i <= top; i++ {
		s += fmt.Sprint(L.ToStringMeta(L.Get(i)).String())
		if i != top {
			s += "\t"
		}
	}
	ctx.luaCtx.Logger().Info(s)
	return 0
}

func lGroup(L *lua.LState) int {
	ctx := upContext(L)
	name := L.CheckString(1)
	fn := L.CheckFunction(2)
	defer ctx.luaCtx.Enter(name).Exit()
	start := time.Now()
	err := L.CallByParam(lua.P{
		Fn:      fn,
		Protect: true,
	})
	cost := time.Since(start)
	scopeName := ctx.luaCtx.Scope()
	metric := luautil.NewGlobalUniqueMetric(ctx.luaCtx.Global(), scopeName, func() metrics.Metric {
		return metrics.Gauge(scopeName+"_us", nil)
	})
	metric.Add(float64(cost.Microseconds()))
	if err != nil {
		L.RaiseError(err.Error())
	}
	return 0
}

func lSleep(L *lua.LState) int {
	ctx := upContext(L)
	dur := time.Duration(L.CheckInt64(1))
	return libasync.Deferred(L, ctx.luaCtx.AsyncPool(), func(ctx context.Context) error {
		time.Sleep(dur)
		return nil
	})
}

func lImport(L *lua.LState) int {
	ctx := upContext(L)
	name := L.CheckString(1)
	L.Pop(L.GetTop())
	b, err := afero.ReadFile(ctx.fs, name)
	if err != nil {
		L.RaiseError(err.Error())
	}
	if err := L.DoString(string(b)); err != nil {
		L.RaiseError(err.Error())
	}
	return L.GetTop()
}
