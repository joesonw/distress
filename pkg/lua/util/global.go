package util

import lua "github.com/yuin/gopher-lua"

func RegisterGlobalFuncs(L *lua.LState, funcs map[string]lua.LGFunction, upvalues ...lua.LValue) {
	for name, f := range funcs {
		L.SetGlobal(name, L.NewClosure(f, upvalues...))
	}
}
