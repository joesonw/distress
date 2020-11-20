package buffer

import (
	"bytes"
	"encoding/binary"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libbytes "github.com/joesonw/distress/pkg/lua/lib/bytes"
)

const moduleName = "buffer"
const metaName = "*BUFFER*"

func Open(L *lua.LState, luaCtx *luacontext.Context) {
	mt := L.NewTypeMetatable(metaName)
	L.SetField(mt, "__index", L.SetFuncs(L.NewTable(), funcs))

	mod := L.RegisterModule(moduleName, map[string]lua.LGFunction{}).(*lua.LTable)
	mod.RawSetString("new", L.NewClosure(func(L *lua.LState) int {
		b := libbytes.Check(L, 2)
		L.Push(New(L, b))
		return 1
	}))
}

func New(L *lua.LState, b []byte) lua.LValue {
	ud := L.NewUserData()
	ud.Value = &bufferContext{buffer: bytes.NewBuffer(b)}
	L.SetMetatable(ud, L.GetTypeMetatable(metaName))
	return ud
}

func Check(L *lua.LState, index int) *bytes.Buffer {
	return CheckValue(L, L.Get(index))
}

func CheckValue(L *lua.LState, val lua.LValue) *bytes.Buffer {
	if val == lua.LNil {
		return nil
	}

	ud, ok := val.(*lua.LUserData)
	if !ok {
		L.RaiseError("expected *BUFFER*")
	}
	c, ok := ud.Value.(*bufferContext)
	if !ok {
		L.RaiseError("expected *BUFFER*")
	}
	return c.buffer
}

type bufferContext struct {
	buffer *bytes.Buffer
}

func upBuffer(L *lua.LState) *bytes.Buffer {
	return L.CheckUserData(1).Value.(*bufferContext).buffer
}

var funcs = map[string]lua.LGFunction{
	"bytes": func(L *lua.LState) int {
		buffer := upBuffer(L)
		L.Push(libbytes.New(L, buffer.Bytes()))
		return 1
	},
	"size": func(L *lua.LState) int {
		buffer := upBuffer(L)
		L.Push(lua.LNumber(buffer.Len()))
		return 1
	},
	"write": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := libbytes.Check(L, 2)
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"read": func(L *lua.LState) int {
		buffer := upBuffer(L)
		size := L.CheckInt(2)
		b := make([]byte, size)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(libbytes.New(L, b))
		return 1
	},
	"write_byte": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := L.CheckNumber(2)
		b := []byte{uint8(n)}
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"read_byte": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 1)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(b[0]))
		return 1
	},
	"write_uint64_be": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := uint64(L.CheckNumber(2))
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(n))
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"write_uint32_be": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := uint32(L.CheckNumber(2))
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(n))
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"write_uint16_be": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := uint16(L.CheckNumber(2))
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(n))
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"write_int64_be": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := int64(L.CheckNumber(2))
		b := make([]byte, 8)
		binary.BigEndian.PutUint64(b, uint64(n))
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"write_int32_be": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := int32(L.CheckNumber(2))
		b := make([]byte, 4)
		binary.BigEndian.PutUint32(b, uint32(n))
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"write_int16_be": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := int16(L.CheckNumber(2))
		b := make([]byte, 2)
		binary.BigEndian.PutUint16(b, uint16(n))
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"write_uint64_le": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := uint64(L.CheckNumber(2))
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(n))
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"write_uint32_le": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := uint32(L.CheckNumber(2))
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(n))
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"write_uint16_le": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := uint16(L.CheckNumber(2))
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(n))
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"write_int64_le": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := int64(L.CheckNumber(2))
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(n))
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"write_int32_le": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := int32(L.CheckNumber(2))
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(n))
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"write_int16_le": func(L *lua.LState) int {
		buffer := upBuffer(L)
		n := int16(L.CheckNumber(2))
		b := make([]byte, 2)
		binary.LittleEndian.PutUint16(b, uint16(n))
		_, err := buffer.Write(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		return 0
	},
	"read_uint64_be": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 8)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(uint64(binary.BigEndian.Uint64(b))))
		return 1
	},
	"read_uint32_be": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 4)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(uint32(binary.BigEndian.Uint32(b))))
		return 1
	},
	"read_uint16_be": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 2)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(uint16(binary.BigEndian.Uint16(b))))
		return 1
	},
	"read_int64_be": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 8)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(int64(binary.BigEndian.Uint64(b))))
		return 1
	},
	"read_int32_be": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 4)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(int32(binary.BigEndian.Uint32(b))))
		return 1
	},
	"read_int16_be": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 2)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(int16(binary.BigEndian.Uint16(b))))
		return 1
	},
	"read_uint64_le": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 8)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(uint64(binary.LittleEndian.Uint64(b))))
		return 1
	},
	"read_uint32_le": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 4)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(uint32(binary.LittleEndian.Uint32(b))))
		return 1
	},
	"read_uint16_le": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 2)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(uint16(binary.LittleEndian.Uint16(b))))
		return 1
	},
	"read_int64_le": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 8)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(int64(binary.LittleEndian.Uint64(b))))
		return 1
	},
	"read_int32_le": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 4)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(int32(binary.LittleEndian.Uint32(b))))
		return 1
	},
	"read_int16_le": func(L *lua.LState) int {
		buffer := upBuffer(L)
		b := make([]byte, 2)
		_, err := buffer.Read(b)
		if err != nil {
			L.RaiseError(err.Error())
		}
		L.Push(lua.LNumber(int16(binary.LittleEndian.Uint16(b))))
		return 1
	},
}
