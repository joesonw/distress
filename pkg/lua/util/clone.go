package util

import lua "github.com/yuin/gopher-lua"

func CloneLuaTable(L *lua.LState, table *lua.LTable) *lua.LTable {
	newTable := L.NewTable()
	table.ForEach(func(k, v lua.LValue) {
		newTable.RawSet(CloneLuaValue(L, k), CloneLuaValue(L, v))
	})
	return newTable
}

func CloneLuaValue(L *lua.LState, value lua.LValue) lua.LValue {
	switch value.Type() {
	case lua.LTNumber, lua.LTBool, lua.LTString:
		return value
	case lua.LTTable:
		return CloneLuaTable(L, value.(*lua.LTable))
	case lua.LTUserData:
		{
			oldUD := value.(*lua.LUserData)
			ud := L.NewUserData()
			ud.Value = oldUD.Value
			return ud
		}
	}
	return lua.LNil
}
