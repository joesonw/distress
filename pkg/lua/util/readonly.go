package util

import lua "github.com/yuin/gopher-lua"

func LRaiseReadOnly(L *lua.LState) int {
	L.RaiseError("Attempt to modify read-only table")
	return 0
}

func NewReadOnly(L *lua.LState, value, toString lua.LValue) lua.LValue {
	proxy := L.NewTable()

	meta := L.NewTable()
	meta.RawSetString("__newindex", L.NewClosure(LRaiseReadOnly))
	meta.RawSetString("__index", value)
	meta.RawSetString("__metadata", lua.LBool(false))
	meta.RawSetString("__tostring", toString)

	L.SetMetatable(proxy, meta)
	return proxy
}
