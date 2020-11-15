package base

import (
	"context"
	"time"

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
	"group": lGroup,
	"sleep": lSleep,
	"fail":  lFail,
}

func lGroup(L *lua.LState) int {
	ctx := upContext(L)
	name := L.CheckString(1)
	fn := L.CheckFunction(2)
	defer ctx.Enter(name).Exit()
	err := L.CallByParam(lua.P{
		Fn:      fn,
		Protect: true,
	})
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
