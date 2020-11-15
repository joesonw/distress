package util

import lua "github.com/yuin/gopher-lua"

func NewOrGetMetadata(L *lua.LState, name string) (*lua.LTable, bool) {
	mt := L.GetTypeMetatable(name)
	if mt != nil {
		return mt.(*lua.LTable), true
	}

	newMt := L.NewTypeMetatable(name)
	newMt.RawSetString("__index", newMt)
	return newMt, false
}
