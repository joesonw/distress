package bytes_test

import (
	"bytes"
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libbytes "github.com/joesonw/distress/pkg/lua/lib/bytes"
	test_util "github.com/joesonw/distress/pkg/lua/test-util"
)

var testTable = []struct {
	name   string
	data   []byte
	script string
	after  test_util.After
}{{
	name: "size",
	data: []byte("hello"),
	script: `
		assert(data:size() == 5, "size")
	`,
}, {
	name: "string",
	data: []byte("hello"),
	script: `
		assert(data:string() == "hello", "string")
	`,
}, {
	name: "string(base64)",
	data: mustDecode(base64.StdEncoding.DecodeString("aGVsbG8=")),
	script: `
		assert(data:string("base64") == "aGVsbG8=", "string(base64)")
	`,
}, {
	name: "string(base32)",
	data: mustDecode(base32.StdEncoding.DecodeString("NBSWY3DP")),
	script: `
		assert(data:string("base32") == "NBSWY3DP", "string(base32)")
	`,
}, {
	name: "string(hex)",
	data: mustDecode(hex.DecodeString("abcdef")),
	script: `
		assert(data:string("hex") == "abcdef", "string(hex)")
	`,
}, {
	name: "string(hex)",
	data: mustDecode(hex.DecodeString("abcdef")),
	script: `
		assert(data:string("hex") == "abcdef", "string(hex)")
	`,
}, {
	name: "get",
	data: []byte{13},
	script: `
		assert(data:get(1) == 13, "string(hex)")
	`,
}, {
	name: "set",
	data: []byte{10},
	script: `
		data:set(2, 20)
	`,
	after: func(t *testing.T, L *lua.LState) {
		b := libbytes.CheckValue(L, L.GetGlobal("data"))
		assert.True(t, bytes.Equal(b, []byte{10, 20}))
	},
}, {
	name: "replace",
	data: []byte("hello"),
	script: `
		data:replace("hello world")
	`,
	after: func(t *testing.T, L *lua.LState) {
		b := libbytes.CheckValue(L, L.GetGlobal("data"))
		assert.True(t, bytes.Equal(b, []byte("hello world")))
	},
}}

func mustDecode(b []byte, err error) []byte {
	if err != nil {
		panic(err)
	}
	return b
}

func Test(t *testing.T) {
	tests := make([]test_util.Testable, len(testTable))
	for i := range testTable {
		test := testTable[i]
		tests[i] = func(_ *testing.T) *test_util.Test {
			return test_util.New(test.name, test.script).
				Before(func(t *testing.T, L *lua.LState, luaCtx *luacontext.Context) {
					libbytes.Open(L, luaCtx)
					L.SetGlobal("data", libbytes.New(L, test.data))
				}).
				After(test.after)
		}
	}
	test_util.Run(t, tests...)
}
