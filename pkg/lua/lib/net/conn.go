package net

import (
	"net"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libgoio "github.com/joesonw/lte/pkg/lua/lib/go-io"
	libpool "github.com/joesonw/lte/pkg/lua/lib/pool"
)

const connMetaName = "*NET*CONN*"

var (
	_ libgoio.Closer = (*connContext)(nil)
	_ libgoio.Writer = (*connContext)(nil)
	_ libgoio.Reader = (*connContext)(nil)
)

type connContext struct {
	net.Conn
	name   string
	guard  *libpool.Guard
	luaCtx *luacontext.Context
}

func (f *connContext) GetName() string {
	return f.name
}

func (f *connContext) GetContext() *luacontext.Context {
	return f.luaCtx
}

func (f *connContext) GetGuard() *libpool.Guard {
	return f.guard
}

var connFuncs = map[string]lua.LGFunction{
	"read":  libgoio.Read,
	"write": libgoio.Write,
	"close": libgoio.Close,
}
