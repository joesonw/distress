package buffer_test

import (
	"encoding/binary"
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libbuffer "github.com/joesonw/lte/pkg/lua/lib/buffer"
	libbytes "github.com/joesonw/lte/pkg/lua/lib/bytes"
	test_util "github.com/joesonw/lte/pkg/lua/test-util"
)

var testTable = []struct {
	name   string
	input  []byte
	script string
	output []byte
}{{
	name:  "bytes",
	input: []byte("hello"),
	script: `
		assert(input:bytes():string() == "hello")
	`,
}, {
	name:  "size",
	input: []byte("hello"),
	script: `
		assert(input:size() == 5)
	`,
}, {
	name:   "write",
	output: []byte("hello"),
	script: `
		local buffer = require "buffer"
		output = buffer:new()
		output:write("hello")
	`,
}, {
	name:  "read",
	input: []byte("hello world"),
	script: `
		assert(input:read(5):string() == "hello")
		assert(input:read(6):string() == " world")
	`,
}, {
	name:   "write_byte",
	output: []byte{20},
	script: `
		local buffer = require "buffer"
		output = buffer:new()
		output:write_byte(20)
	`,
}, {
	name:  "read_byte",
	input: []byte{20},
	script: `
		assert(input:read_byte() == 20)
	`,
}, {
	name: "write_uint64_be",
	output: writeBytes(8, func(b []byte) {
		i := uint64(240)
		binary.BigEndian.PutUint64(b, uint64(i))
	}),
	script: `
        local buffer = require "buffer"
        output = buffer:new()
        output:write_uint64_be(240)
    `,
}, {
	name: "write_uint32_be",
	output: writeBytes(4, func(b []byte) {
		i := uint32(240)
		binary.BigEndian.PutUint32(b, uint32(i))
	}),
	script: `
        local buffer = require "buffer"
        output = buffer:new()
        output:write_uint32_be(240)
    `,
}, {
	name: "write_uint16_be",
	output: writeBytes(2, func(b []byte) {
		i := uint16(240)
		binary.BigEndian.PutUint16(b, uint16(i))
	}),
	script: `
        local buffer = require "buffer"
        output = buffer:new()
        output:write_uint16_be(240)
    `,
}, {
	name: "write_int64_be",
	output: writeBytes(8, func(b []byte) {
		i := int64(-240)
		binary.BigEndian.PutUint64(b, uint64(i))
	}),
	script: `
        local buffer = require "buffer"
        output = buffer:new()
        output:write_int64_be(-240)
    `,
}, {
	name: "write_int32_be",
	output: writeBytes(4, func(b []byte) {
		i := int32(-240)
		binary.BigEndian.PutUint32(b, uint32(i))
	}),
	script: `
        local buffer = require "buffer"
        output = buffer:new()
        output:write_int32_be(-240)
    `,
}, {
	name: "write_int16_be",
	output: writeBytes(2, func(b []byte) {
		i := int16(-240)
		binary.BigEndian.PutUint16(b, uint16(i))
	}),
	script: `
        local buffer = require "buffer"
        output = buffer:new()
        output:write_int16_be(-240)
    `,
}, {
	name: "write_uint64_le",
	output: writeBytes(8, func(b []byte) {
		i := uint64(240)
		binary.LittleEndian.PutUint64(b, uint64(i))
	}),
	script: `
        local buffer = require "buffer"
        output = buffer:new()
        output:write_uint64_le(240)
    `,
}, {
	name: "write_uint32_le",
	output: writeBytes(4, func(b []byte) {
		i := uint32(240)
		binary.LittleEndian.PutUint32(b, uint32(i))
	}),
	script: `
        local buffer = require "buffer"
        output = buffer:new()
        output:write_uint32_le(240)
    `,
}, {
	name: "write_uint16_le",
	output: writeBytes(2, func(b []byte) {
		i := uint16(240)
		binary.LittleEndian.PutUint16(b, uint16(i))
	}),
	script: `
        local buffer = require "buffer"
        output = buffer:new()
        output:write_uint16_le(240)
    `,
}, {
	name: "write_int64_le",
	output: writeBytes(8, func(b []byte) {
		i := int64(-240)
		binary.LittleEndian.PutUint64(b, uint64(i))
	}),
	script: `
        local buffer = require "buffer"
        output = buffer:new()
        output:write_int64_le(-240)
    `,
}, {
	name: "write_int32_le",
	output: writeBytes(4, func(b []byte) {
		i := int32(-240)
		binary.LittleEndian.PutUint32(b, uint32(i))
	}),
	script: `
        local buffer = require "buffer"
        output = buffer:new()
        output:write_int32_le(-240)
    `,
}, {
	name: "write_int16_le",
	output: writeBytes(2, func(b []byte) {
		i := int16(-240)
		binary.LittleEndian.PutUint16(b, uint16(i))
	}),
	script: `
        local buffer = require "buffer"
        output = buffer:new()
        output:write_int16_le(-240)
    `,
}, {
	name: "read_uint64_be",
	input: writeBytes(8, func(b []byte) {
		i := uint64(240)
		binary.BigEndian.PutUint64(b, uint64(i))
	}),
	script: `
        assert(input:read_uint64_be() == 240)
    `,
}, {
	name: "read_uint32_be",
	input: writeBytes(4, func(b []byte) {
		i := uint32(240)
		binary.BigEndian.PutUint32(b, uint32(i))
	}),
	script: `
        assert(input:read_uint32_be() == 240)
    `,
}, {
	name: "read_uint16_be",
	input: writeBytes(2, func(b []byte) {
		i := uint16(240)
		binary.BigEndian.PutUint16(b, uint16(i))
	}),
	script: `
        assert(input:read_uint16_be() == 240)
    `,
}, {
	name: "read_int64_be",
	input: writeBytes(8, func(b []byte) {
		i := int64(-240)
		binary.BigEndian.PutUint64(b, uint64(i))
	}),
	script: `
        assert(input:read_int64_be() == -240)
    `,
}, {
	name: "read_int32_be",
	input: writeBytes(4, func(b []byte) {
		i := int32(-240)
		binary.BigEndian.PutUint32(b, uint32(i))
	}),
	script: `
        assert(input:read_int32_be() == -240)
    `,
}, {
	name: "read_int16_be",
	input: writeBytes(2, func(b []byte) {
		i := int16(-240)
		binary.BigEndian.PutUint16(b, uint16(i))
	}),
	script: `
        assert(input:read_int16_be() == -240)
    `,
}, {
	name: "read_uint64_le",
	input: writeBytes(8, func(b []byte) {
		i := uint64(240)
		binary.LittleEndian.PutUint64(b, uint64(i))
	}),
	script: `
        assert(input:read_uint64_le() == 240)
    `,
}, {
	name: "read_uint32_le",
	input: writeBytes(4, func(b []byte) {
		i := uint32(240)
		binary.LittleEndian.PutUint32(b, uint32(i))
	}),
	script: `
        assert(input:read_uint32_le() == 240)
    `,
}, {
	name: "read_uint16_le",
	input: writeBytes(2, func(b []byte) {
		i := uint16(240)
		binary.LittleEndian.PutUint16(b, uint16(i))
	}),
	script: `
        assert(input:read_uint16_le() == 240)
    `,
}, {
	name: "read_int64_le",
	input: writeBytes(8, func(b []byte) {
		i := int64(-240)
		binary.LittleEndian.PutUint64(b, uint64(i))
	}),
	script: `
        assert(input:read_int64_le() == -240)
    `,
}, {
	name: "read_int32_le",
	input: writeBytes(4, func(b []byte) {
		i := int32(-240)
		binary.LittleEndian.PutUint32(b, uint32(i))
	}),
	script: `
        assert(input:read_int32_le() == -240)
    `,
}, {
	name: "read_int16_le",
	input: writeBytes(2, func(b []byte) {
		i := int16(-240)
		binary.LittleEndian.PutUint16(b, uint16(i))
	}),
	script: `
        assert(input:read_int16_le() == -240)
    `,
}}

func writeBytes(size int, f func([]byte)) []byte {
	b := make([]byte, size)
	f(b)
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
					libbuffer.Open(L, luaCtx)
					L.SetGlobal("input", libbuffer.New(L, test.input))
				}).
				After(func(t *testing.T, L *lua.LState) {
					if test.output != nil {
						b := libbuffer.CheckValue(L, L.GetGlobal("output"))
						assert.Equal(t, string(test.output), b.String())
					}
				})
		}
	}
	test_util.Run(t, tests...)
}
