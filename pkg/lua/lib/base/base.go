package base

import (
	"context"
	"fmt"
	"time"

	"github.com/joesonw/distress/pkg/metrics"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libasync "github.com/joesonw/distress/pkg/lua/lib/async"
)

func Open(L *lua.LState, luaCtx *luacontext.Context) {
	ud := L.NewUserData()
	ud.Value = luaCtx
	for k, f := range funcs {
		L.SetGlobal(k, L.NewClosure(f, ud))
	}
}

func upContext(L *lua.LState) *luacontext.Context {
	ud := L.CheckUserData(lua.UpvalueIndex(1))
	return ud.Value.(*luacontext.Context)
}

var funcs = map[string]lua.LGFunction{
	"print": lPrint,
	"group": lGroup,
	"sleep": lSleep,
	"fail":  lFail,
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
	ctx.Logger().Info(s)
	return 0
}

func lGroup(L *lua.LState) int {
	ctx := upContext(L)
	name := L.CheckString(1)
	fn := L.CheckFunction(2)
	defer ctx.Enter(name).Exit()
	start := time.Now()
	err := L.CallByParam(lua.P{
		Fn:      fn,
		Protect: true,
	})
	cost := time.Since(start)
	scopeName := ctx.Scope()
	in := ctx.Global().Unique(scopeName, func() interface{} {
		metric := metrics.Gauge(scopeName+"_us", nil)
		ctx.Global().RegisterMetric(metric)
		return metric
	})
	ctx.Global()
	metric := in.(metrics.Metric)
	metric.Add(float64(cost.Microseconds()))
	if err != nil {
		L.RaiseError(err.Error())
	}
	return 0
}

func lSleep(L *lua.LState) int {
	ctx := upContext(L)
	dur := time.Duration(L.CheckInt64(2))
	return libasync.Deferred(L, ctx.AsyncPool(), func(ctx context.Context) error {
		time.Sleep(dur)
		return nil
	})
}

func lFail(L *lua.LState) int {
	err := L.Get(1)
	L.RaiseError(err.String())
	return 0
}
