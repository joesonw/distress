package uuid

import (
	uuid "github.com/satori/go.uuid"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
)

const moduleName = "uuid"

func Open(L *lua.LState, luaCtx *luacontext.Context) {
	mod := L.RegisterModule(moduleName, map[string]lua.LGFunction{}).(*lua.LTable)

	mod.RawSetString("v5", L.NewClosure(v5))
	mod.RawSetString("v4", L.NewClosure(v4))
	mod.RawSetString("v3", L.NewClosure(v3))
	mod.RawSetString("v2", L.NewClosure(v2))
	mod.RawSetString("v1", L.NewClosure(v1))
}

func v5(L *lua.LState) int {
	ns := L.CheckString(2)
	name := L.CheckString(3)
	L.Push(lua.LString(uuid.NewV5(uuid.FromStringOrNil(ns), name).String()))
	return 1
}

func v4(L *lua.LState) int {
	L.Push(lua.LString(uuid.NewV4().String()))
	return 1
}

func v3(L *lua.LState) int {
	ns := L.CheckString(2)
	name := L.CheckString(3)
	L.Push(lua.LString(uuid.NewV3(uuid.FromStringOrNil(ns), name).String()))
	return 1
}

func v2(L *lua.LState) int {
	domain := L.CheckInt(2)
	L.Push(lua.LString(uuid.NewV2(byte(domain)).String()))
	return 1
}

func v1(L *lua.LState) int {
	L.Push(lua.LString(uuid.NewV1().String()))
	return 1
}
