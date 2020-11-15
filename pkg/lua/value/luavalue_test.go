package value_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"

	luavalue "github.com/joesonw/distress/pkg/lua/value"
)

func TestLuaTableToStruct(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	table := L.NewTable()
	table.RawSetString("int64", lua.LNumber(10))
	table.RawSetString("float", lua.LNumber(10.1))
	table.RawSetString("string", lua.LString("hello"))
	table.RawSetString("bool", lua.LBool(true))
	arr := L.NewTable()
	for i := 0; i < 3; i++ {
		arr.RawSetInt(i+1, lua.LNumber(i+1))
	}
	table.RawSetString("arr", arr)

	type S struct {
		Int64  int64   `json:"int64,omitempty"`
		Float  float64 `json:"float,omitempty"`
		String string  `json:"string,omitempty"`
		Bool   bool    `json:"bool,omitempty"`
		Arr    []int64 `json:"arr,omitempty"`
	}

	s := S{}
	assert.Nil(t, luavalue.Unmarshal(table, &s))
	assert.Equal(t, int64(10), s.Int64)
	assert.Equal(t, float64(10.1), s.Float)
	assert.Equal(t, "hello", s.String)
	assert.Equal(t, true, s.Bool)
	assert.ElementsMatch(t, []int64{1, 2, 3}, s.Arr)
}

func TestLuaStructToTable(t *testing.T) {
	L := lua.NewState()
	defer L.Close()

	type S struct {
		Int64  int64   `json:"int64,omitempty"`
		Float  float64 `json:"float,omitempty"`
		String string  `json:"string,omitempty"`
		Bool   bool    `json:"bool,omitempty"`
		Arr    []int64 `json:"arr,omitempty"`
	}

	s := S{
		Int64:  10,
		Float:  10.2,
		String: "hello",
		Bool:   true,
		Arr:    []int64{1, 2, 3},
	}

	in, err := luavalue.Marshal(L, &s)
	assert.Nil(t, err)
	assert.NotNil(t, in)
	table := in.(*lua.LTable)
	assert.Equal(t, lua.LNumber(10), table.RawGetString("int64").(lua.LNumber))
	assert.Equal(t, lua.LNumber(10.2), table.RawGetString("float").(lua.LNumber))
	assert.Equal(t, lua.LString("hello"), table.RawGetString("string").(lua.LString))
	assert.Equal(t, lua.LBool(true), table.RawGetString("bool").(lua.LBool))
	assert.Equal(t, lua.LNumber(1), table.RawGetString("arr").(*lua.LTable).RawGetInt(1).(lua.LNumber))
	assert.Equal(t, lua.LNumber(2), table.RawGetString("arr").(*lua.LTable).RawGetInt(2).(lua.LNumber))
	assert.Equal(t, lua.LNumber(3), table.RawGetString("arr").(*lua.LTable).RawGetInt(3).(lua.LNumber))
}
