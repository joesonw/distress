package util

import (
	lua "github.com/yuin/gopher-lua"
)

func EnsureParamTypes(L *lua.LState, types ...lua.LValueType) bool {
	n := L.GetTop()
	if n != len(types) {
		L.RaiseError("require %d arguments, has %d", len(types), n)
		return true
	}

	for i := 0; i < n; i++ {
		val := L.Get(i + 1)
		if val.Type() != types[i] {
			L.TypeError(i+1, types[i])
			return false
		}
	}

	return true
}

func EnsureTableProperties(L *lua.LState, table *lua.LTable, properties map[string]lua.LValueType) bool {
	for key, typ := range properties {
		val := table.RawGetString(key)
		if val.Type() != typ {
			L.RaiseError("key \"%s\" of table should be %s", key, typ.String())
			return false
		}
	}

	return true
}
