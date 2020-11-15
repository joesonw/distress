package fs_test

import (
	"io/ioutil"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libfs "github.com/joesonw/distress/pkg/lua/lib/fs"
	test_util "github.com/joesonw/distress/pkg/lua/test-util"
)

var testTable = []struct {
	name   string
	files  map[string]string
	script string
	after  func(t *testing.T, fs afero.Fs)
}{{
	name: "fs:open",
	files: map[string]string{
		"/test": "hello world",
	},
	script: `
		local fs = require "fs"
		local err, file = fs:open("/test")()
		assert(err == nil, "open file")
		assert(file:name() == "/test", "name")

		local err, contents = file:read(5)()
		assert(err == nil, "read file")
		assert(contents:string() == "hello", "read file")
		local err = file:close()()
		assert(err == nil, "close file")

		local _, file = fs:open("/test")()
		local err, contents = file:read_all()()
		assert(err == nil, "read all")
		assert(contents:string() == "hello world", "read all")
		local err = file:close()()
		assert(err == nil, "close file")

		local _, file = fs:open("/test")()
		local err = file:write("nihao")()
		assert(err == nil, "write")
		local err = file:close()()
		assert(err == nil, "close file")
	`,
	after: func(t *testing.T, fs afero.Fs) {
		f, err := fs.Open("/test")
		assert.Nil(t, err)
		b, err := ioutil.ReadAll(f)
		assert.Nil(t, err)
		assert.Equal(t, "nihao world", string(b))
	},
}, {
	name: "fs:create",
	script: `
		local fs = require "fs"
		local err, file = fs:create("/test")()
		assert(err == nil, "create file")
		local err = file:write("nihao")()
		assert(err == nil, "write file")
		local err = file:close()()
		assert(err == nil, "close file")
	`,
	after: func(t *testing.T, fs afero.Fs) {
		f, err := fs.Open("/test")
		assert.Nil(t, err)
		b, err := ioutil.ReadAll(f)
		assert.Nil(t, err)
		assert.Equal(t, "nihao", string(b))
	},
}, {
	name:  "fs:remove",
	files: map[string]string{"/test": "123"},
	script: `
		local fs = require "fs"
		local err = fs:remove("/test")()
		assert(err == nil, "fs remove")
	`,
	after: func(t *testing.T, fs afero.Fs) {
		exists, err := afero.Exists(fs, "/test")
		assert.Nil(t, err)
		assert.False(t, exists)
	},
}, {
	name: "fs:remove_all",
	files: map[string]string{
		"/test/a": "123",
		"/test/b": "123",
	},
	script: `
		local fs = require "fs"
		local err = fs:remove_all("/test")()
		assert(err == nil, "fs remove all")
	`,
	after: func(t *testing.T, fs afero.Fs) {
		exists, err := afero.Exists(fs, "/test/123")
		assert.Nil(t, err)
		assert.False(t, exists)
	},
}, {
	name: "fs:list",
	files: map[string]string{
		"/test/a": "aaa",
		"/test/b": "bbb",
	},
	script: `
		local fs = require "fs"
		local err, list = fs:list("/test")()
		assert(err == nil, "fs list")
		assert(list[1].name == "a", "fs list")
		assert(list[1].size == 3, "fs list")
		assert(list[1].dir == false, "fs list")
		assert(list[2].name == "b", "fs list")
		assert(list[2].size == 3, "fs list")
		assert(list[2].dir == false, "fs list")
	`,
}, {
	name:  "fs:stat",
	files: map[string]string{"/test": "123"},
	script: `
		local fs = require "fs"
		local err, stat = fs:stat("/test")()
		assert(err == nil, "fs stat")
		assert(stat.name == "test", "fs stat")
		assert(stat.size == 3, "fs stat")
		assert(stat.dir == false, "fs stat")
	`,
}, {
	name: "fs:mkdir_all",
	script: `
		local fs = require "fs"
		local err = fs:mkdir_all("/test/123")()
		assert(err == nil, "fs mkdir all")
	`,
	after: func(t *testing.T, fs afero.Fs) {
		exists, err := afero.DirExists(fs, "/test/123")
		assert.Nil(t, err)
		assert.True(t, exists)
	},
}, {
	name:  "fs:exist",
	files: map[string]string{"/test": "123"},
	script: `
		local fs = require "fs"
		local err, ok = fs:exist("/test")()
		assert(err == nil, "fs exist")
		assert(ok == true, "fs exist")

		local err, ok = fs:exist("/test/2")()
		assert(err == nil, "fs exist")
		assert(ok == false, "fs exist")
	`,
}, {
	name:  "fs:read_all",
	files: map[string]string{"/test": "123"},
	script: `
		local fs = require "fs"
		local err, contents = fs:read_all("/test")()
		assert(err == nil, "fs read all")
		assert(contents:string() == "123", "fs read all")
	`,
}, {
	name: "fs:write",
	script: `
		local fs = require "fs"
		local err = fs:write("/test", "123")()
		assert(err == nil, "fs write")
	`,
	after: func(t *testing.T, fs afero.Fs) {
		f, err := fs.Open("/test")
		assert.Nil(t, err)
		b, err := afero.ReadAll(f)
		assert.Nil(t, err)
		assert.Equal(t, string(b), "123")
	},
}}

func Test(t *testing.T) {
	tests := make([]test_util.Testable, len(testTable))
	for i := range testTable {
		test := testTable[i]
		tests[i] = func(_ *testing.T) *test_util.Test {
			fs := afero.NewMemMapFs()

			for name, contents := range test.files {
				f, _ := fs.Create(name)
				_, _ = f.WriteString(contents)
			}

			return test_util.New(test.name, test.script).
				Before(func(t *testing.T, L *lua.LState, luaCtx *luacontext.Context) {
					libfs.Open(L, luaCtx, fs)
				}).
				After(func(t *testing.T, L *lua.LState) {
					if test.after != nil {
						test.after(t, fs)
					}
				})
		}
	}
	test_util.Run(t, tests...)
}
