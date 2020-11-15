package goclass

import (
	lua "github.com/yuin/gopher-lua"
)

type Class struct {
	metatable *lua.LTable
	name      string
}

func (c *Class) New(L *lua.LState, value interface{}) lua.LValue {
	ud := L.NewUserData()
	ud.Value = value
	L.SetMetatable(ud, c.metatable)
	return ud
}

func New(L *lua.LState, name string, methods map[string]lua.LGFunction) *Class {
	mt := L.NewTypeMetatable(name)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))
	return &Class{
		name:      name,
		metatable: mt,
	}
}
