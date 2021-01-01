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

func (c *Class) WithIndex(v lua.LValue) *Class {
	c.metatable.RawSetString("__index", v)
	return c
}

func (c *Class) WithNewIndex(v lua.LValue) *Class {
	c.metatable.RawSetString("__newindex", v)
	return c
}

func (c *Class) WithCall(v lua.LValue) *Class {
	c.metatable.RawSetString("__call", v)
	return c
}

func (c *Class) WithToString(v lua.LValue) *Class {
	c.metatable.RawSetString("__tostring", v)
	return c
}

func (c *Class) WithUnaryMinus(v lua.LValue) *Class {
	c.metatable.RawSetString("__unm", v)
	return c
}

func (c *Class) WithAdd(v lua.LValue) *Class {
	c.metatable.RawSetString("__add", v)
	return c
}

func (c *Class) WithSub(v lua.LValue) *Class {
	c.metatable.RawSetString("__sub", v)
	return c
}

func (c *Class) WithMul(v lua.LValue) *Class {
	c.metatable.RawSetString("__mul", v)
	return c
}

func (c *Class) WithDiv(v lua.LValue) *Class {
	c.metatable.RawSetString("__div", v)
	return c
}

func (c *Class) WithIDiv(v lua.LValue) *Class {
	c.metatable.RawSetString("__idiv", v)
	return c
}

func (c *Class) WithMod(v lua.LValue) *Class {
	c.metatable.RawSetString("__mod", v)
	return c
}

func (c *Class) WithPow(v lua.LValue) *Class {
	c.metatable.RawSetString("__pow", v)
	return c
}

func (c *Class) WithConcat(v lua.LValue) *Class {
	c.metatable.RawSetString("__concat", v)
	return c
}

func (c *Class) WithAnd(v lua.LValue) *Class {
	c.metatable.RawSetString("__band", v)
	return c
}

func (c *Class) WithOr(v lua.LValue) *Class {
	c.metatable.RawSetString("__bor", v)
	return c
}

func (c *Class) WithXor(v lua.LValue) *Class {
	c.metatable.RawSetString("__bxor", v)
	return c
}

func (c *Class) WithNot(v lua.LValue) *Class {
	c.metatable.RawSetString("__bnot", v)
	return c
}

func (c *Class) WithShl(v lua.LValue) *Class {
	c.metatable.RawSetString("__shl", v)
	return c
}

func (c *Class) WithShr(v lua.LValue) *Class {
	c.metatable.RawSetString("__shr", v)
	return c
}

func (c *Class) WithEq(v lua.LValue) *Class {
	c.metatable.RawSetString("__eq", v)
	return c
}

func (c *Class) WithLt(v lua.LValue) *Class {
	c.metatable.RawSetString("__lt", v)
	return c
}

func (c *Class) WithLe(v lua.LValue) *Class {
	c.metatable.RawSetString("__le", v)
	return c
}

func New(L *lua.LState, name string, methods map[string]lua.LGFunction) *Class {
	mt := L.NewTypeMetatable(name)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), methods))
	return &Class{
		name:      name,
		metatable: mt,
	}
}
