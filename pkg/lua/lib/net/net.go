package net

import (
	"context"
	"net"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libasync "github.com/joesonw/distress/pkg/lua/lib/async"
	goclass "github.com/joesonw/distress/pkg/lua/lib/go-class"
	libpool "github.com/joesonw/distress/pkg/lua/lib/pool"
)

type netContext struct {
	protocol string
	class    *goclass.Class
	luaCtx   *luacontext.Context
}

func open(L *lua.LState, luaCtx *luacontext.Context, protocol string, class *goclass.Class) lua.LValue {
	ud := L.NewUserData()
	ud.Value = &netContext{
		protocol: protocol,
		class:    class,
		luaCtx:   luaCtx,
	}
	return L.NewClosure(netOpen, ud)
}

func Open(L *lua.LState, luaCtx *luacontext.Context) {
	class := goclass.New(L, connMetaName, connFuncs)
	mod := L.RegisterModule("net", map[string]lua.LGFunction{}).(*lua.LTable)
	mod.RawSetString("tcp", open(L, luaCtx, "tcp", class))
	mod.RawSetString("udp", open(L, luaCtx, "udp", class))
}

func netOpen(L *lua.LState) int {
	c := L.CheckUserData(lua.UpvalueIndex(1)).Value.(*netContext)
	addr := L.CheckString(2)

	return libasync.DeferredResult(L, c.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		conn, err := net.Dial(c.protocol, addr)
		if err != nil {
			return nil, err
		}

		g := c.luaCtx.ReleasePool().Watch(libpool.NewIOReadWriteCloserResource(c.protocol+" conn", conn))
		return func(L *lua.LState) int {
			L.Push(c.class.New(L, &connContext{
				Conn:   conn,
				guard:  g,
				luaCtx: c.luaCtx,
			}))
			return 1
		}, nil
	})
}
