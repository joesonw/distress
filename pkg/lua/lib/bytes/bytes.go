package bytes

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
)

const metaName = "*BYTES*"
const moduleName = "bytes"

func Open(L *lua.LState, luaCtx *luacontext.Context) {
	mt := L.NewTypeMetatable(metaName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), funcs))

	mod := L.RegisterModule(moduleName, map[string]lua.LGFunction{}).(*lua.LTable)
	mod.RawSetString("new", L.NewClosure(func(L *lua.LState) int {
		s := L.CheckString(2)
		encoding := ""
		if e := L.Get(3); e != lua.LNil {
			encoding = e.String()
		}

		switch encoding {
		case "base64":
			if b, err := base64.StdEncoding.DecodeString(s); err != nil {
				L.RaiseError(err.Error())
			} else {
				L.Push(New(L, b))
			}
		case "base32":
			if b, err := base32.StdEncoding.DecodeString(s); err != nil {
				L.RaiseError(err.Error())
			} else {
				L.Push(New(L, b))
			}
		case "hex":
			if b, err := hex.DecodeString(s); err != nil {
				L.RaiseError(err.Error())
			} else {
				L.Push(New(L, b))
			}
		default:
			L.Push(New(L, []byte(s)))
		}

		return 1
	}))
}

type context struct {
	bytes []byte
}

func Check(L *lua.LState, index int) []byte {
	return CheckValue(L, L.Get(index))
}

func Is(val lua.LValue) bool {
	if _, ok := val.(lua.LString); ok {
		return true
	}

	ud, ok := val.(*lua.LUserData)
	if !ok {
		return false
	}

	_, ok = ud.Value.(*context)
	return ok
}

func Get(val lua.LValue) []byte {
	if _, ok := val.(lua.LString); ok {
		return []byte(val.String())
	}

	ud, ok := val.(*lua.LUserData)
	if !ok {
		return nil
	}

	c, ok := ud.Value.(*context)
	if !ok {
		return nil
	}

	return c.bytes
}

func CheckValue(L *lua.LState, val lua.LValue) []byte {
	if s, ok := val.(lua.LString); ok {
		return []byte(s)
	}

	ud, ok := val.(*lua.LUserData)
	if !ok {
		L.RaiseError("expected *BYTES* or string")
	}
	c, ok := ud.Value.(*context)
	if !ok {
		L.RaiseError("expected *BYTES*")
	}
	return c.bytes
}

func New(L *lua.LState, bytes []byte) lua.LValue {
	ud := L.NewUserData()
	ud.Value = &context{
		bytes: bytes,
	}
	L.SetMetatable(ud, L.GetTypeMetatable(metaName))
	return ud
}

func upContext(L *lua.LState) *context {
	return L.CheckUserData(1).Value.(*context)
}

var funcs = map[string]lua.LGFunction{
	"size": func(L *lua.LState) int {
		ctx := upContext(L)
		L.Push(lua.LNumber(len(ctx.bytes)))
		return 1
	},
	"string": func(L *lua.LState) int {
		ctx := upContext(L)
		encoding := L.Get(2).String()
		switch encoding {
		case "base64":
			L.Push(lua.LString(base64.StdEncoding.EncodeToString(ctx.bytes)))
		case "base32":
			L.Push(lua.LString(base32.StdEncoding.EncodeToString(ctx.bytes)))
		case "hex":
			L.Push(lua.LString(hex.EncodeToString(ctx.bytes)))
		default:
			L.Push(lua.LString(ctx.bytes))
		}
		return 1
	},
	"get": func(L *lua.LState) int {
		ctx := upContext(L)
		index := L.CheckInt(2)
		L.Push(lua.LNumber(ctx.bytes[index-1]))
		return 1
	},
	"set": func(L *lua.LState) int {
		ctx := upContext(L)
		index := L.CheckInt(2)
		b := L.CheckNumber(3)
		if index >= len(ctx.bytes) {
			newBytes := make([]byte, index)
			copy(newBytes, ctx.bytes)
			ctx.bytes = newBytes
		}
		ctx.bytes[index-1] = byte(b)
		return 0
	},
	"replace": func(L *lua.LState) int {
		ctx := upContext(L)
		s := L.CheckString(2)
		ctx.bytes = []byte(s)
		return 0
	},
}
