package stat

import (
	"time"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	goclass "github.com/joesonw/lte/pkg/lua/lib/go-class"
	stat "github.com/joesonw/lte/pkg/stat"
)

const statMetaName = "*STAT*"
const moduleName = "stats"

type modContext struct {
	luaCtx *luacontext.Context
	class  *goclass.Class
}

type context struct {
	luaCtx *luacontext.Context
	stat   *stat.Stat
}

func Open(L *lua.LState, luaCtx *luacontext.Context) {
	mod := L.RegisterModule(moduleName, map[string]lua.LGFunction{}).(*lua.LTable)
	ud := L.NewUserData()
	ud.Value = &modContext{
		luaCtx: luaCtx,
		class: goclass.New(L, statMetaName, map[string]lua.LGFunction{
			"tag":      lTag,
			"field":    lField,
			"set_time": lSetTime,
			"submit":   lSubmit,
		}),
	}
	L.SetFuncs(mod, map[string]lua.LGFunction{
		"new": func(L *lua.LState) int {
			c := L.CheckUserData(lua.UpvalueIndex(1)).Value.(*modContext)
			name := L.CheckString(2)
			v := c.class.New(L, &context{
				luaCtx: c.luaCtx,
				stat:   stat.New(name),
			})
			L.Push(v)
			return 1
		},
	}, ud)
}

func check(L *lua.LState) *context {
	c := L.CheckUserData(1).Value.(*context)
	return c
}

func lTag(L *lua.LState) int {
	c := check(L)
	name := L.CheckString(2)
	value := L.CheckString(3)
	c.stat.Tag(name, value)
	return 0
}

func lField(L *lua.LState) int {
	c := check(L)
	name := L.CheckString(2)
	value := L.CheckNumber(3)
	c.stat.FloatField(name, float64(value))
	return 0
}

func lSetTime(L *lua.LState) int {
	c := check(L)
	u := L.CheckUserData(2)
	t := u.Value.(time.Time)
	c.stat.SetTime(t)
	return 0
}

func lSubmit(L *lua.LState) int {
	c := check(L)
	c.luaCtx.Global().Report(c.stat)
	return 0
}
