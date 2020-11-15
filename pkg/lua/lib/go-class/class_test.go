package goclass_test

import (
	"sync/atomic"
	"testing"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	goclass "github.com/joesonw/distress/pkg/lua/lib/go-class"
	test_util "github.com/joesonw/distress/pkg/lua/test-util"
)

func Test(t *testing.T) {
	test_util.Run(t, func(t *testing.T) *test_util.Test {
		var id int64
		return test_util.New("go class", `
			local a = create()
			local b = create()
			assert(a:id() == 1, "a")
			assert(b:id() == 2, "b")
		`).Before(func(t *testing.T, L *lua.LState, luaCtx *luacontext.Context) {
			class := goclass.New(L, "*TEST*", map[string]lua.LGFunction{
				"id": func(L *lua.LState) int {
					id := L.CheckUserData(1).Value.(int64)
					L.Push(lua.LNumber(id))
					return 1
				},
			})
			L.SetGlobal("create", L.NewClosure(func(L *lua.LState) int {
				L.Push(class.New(L, atomic.AddInt64(&id, 1)))
				return 1
			}))
		})
	})
}
