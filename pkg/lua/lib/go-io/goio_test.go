package goio_test

import (
	"bytes"
	"io"
	"sync"
	"testing"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	goclass "github.com/joesonw/lte/pkg/lua/lib/go-class"
	libgoio "github.com/joesonw/lte/pkg/lua/lib/go-io"
	libpool "github.com/joesonw/lte/pkg/lua/lib/pool"
	test_util "github.com/joesonw/lte/pkg/lua/test-util"
)

var (
	_ libgoio.Reader = (*readWriter)(nil)
	_ libgoio.Writer = (*readWriter)(nil)
)

type readWriter struct {
	io.ReadWriter
	luaCtx *luacontext.Context
	guard  *libpool.Guard
	closed *sync.WaitGroup
}

func (c *readWriter) GetContext() *luacontext.Context {
	return c.luaCtx
}

func (c *readWriter) GetGuard() *libpool.Guard {
	return c.guard
}

func (c *readWriter) GetName() string {
	return ""
}

func (c *readWriter) Close() error {
	c.closed.Done()
	return nil
}

func Test(t *testing.T) {
	test_util.Run(t, func(t *testing.T) *test_util.Test {
		closed := &sync.WaitGroup{}
		return test_util.New("go io", `
			local f = create()
			local err, read = f:read(5)()
			assert(err == nil, "io:read")
			assert(read:string() == "hello", "io:read")
			local err = f:close()()
			assert(err  == nil, "io:close")	

			local f = create()
			local err, read = f:read_all()()
			assert(err  == nil, "io:read_all")
			assert(read:string() == "hello world", "io:read_all")

			local err = f:write("world hello")()
			assert(err  == nil, "io:write")
			local err, read = f:read_all()()
			assert(err  == nil, "io:write")
			assert(read:string() == "world hello", "io:write")
		
			local err = f:close()()
			assert(err  == nil, "io:close")	
			
		`).Before(func(t *testing.T, L *lua.LState, luaCtx *luacontext.Context) {
			class := goclass.New(L, "*TEST*", map[string]lua.LGFunction{
				"read":     libgoio.Read,
				"read_all": libgoio.ReadAll,
				"write":    libgoio.Write,
				"close":    libgoio.Close,
			})
			L.SetGlobal("create", L.NewClosure(func(L *lua.LState) int {
				L.Push(class.New(L, &readWriter{
					ReadWriter: bytes.NewBufferString("hello world"),
					luaCtx:     luaCtx,
					closed:     closed,
				}))
				closed.Add(1)
				return 1
			}))
		}).After(func(t *testing.T, L *lua.LState) {
			closed.Wait()
		})
	})
}
