package fs

import (
	"github.com/spf13/afero"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libgoio "github.com/joesonw/lte/pkg/lua/lib/go-io"
	libpool "github.com/joesonw/lte/pkg/lua/lib/pool"
)

const fileMetaName = "*FILE*"

var (
	_ libgoio.Reader = (*file)(nil)
	_ libgoio.Writer = (*file)(nil)
	_ libgoio.Closer = (*file)(nil)
)

type file struct {
	afero.File
	name  string
	ctx   *fsContext
	guard *libpool.Guard
}

func (f *file) Name() string {
	return f.name
}

func (f *file) GetContext() *luacontext.Context {
	return f.ctx.luaCtx
}

func (f *file) GetGuard() *libpool.Guard {
	return f.guard
}

type Named interface {
	Name() string
}

var fileFuncs = map[string]lua.LGFunction{
	"read":     libgoio.Read,
	"read_all": libgoio.ReadAll,
	"write":    libgoio.Write,
	"close":    libgoio.Close,
	"name": func(L *lua.LState) int {
		ud := L.CheckUserData(1)
		named := ud.Value.(Named)
		L.Push(lua.LString(named.Name()))
		return 1
	},
}
