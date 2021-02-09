package websocket

import (
	"context"

	"github.com/gobwas/ws"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libasync "github.com/joesonw/lte/pkg/lua/lib/async"
	goclass "github.com/joesonw/lte/pkg/lua/lib/go-class"
	libpool "github.com/joesonw/lte/pkg/lua/lib/pool"
)

const moduleName = "websocket"

type wsContext struct {
	class  *goclass.Class
	luaCtx *luacontext.Context
}

func Open(L *lua.LState, luaCtx *luacontext.Context) {
	mod := L.RegisterModule(moduleName, map[string]lua.LGFunction{}).(*lua.LTable)

	ud := L.NewUserData()
	ud.Value = &wsContext{
		luaCtx: luaCtx,
		class:  goclass.New(L, connMetaName, connFuncs),
	}
	mod.RawSetString("open", L.NewClosure(wsOpen, ud))
}

func wsOpen(L *lua.LState) int {
	c := L.CheckUserData(lua.UpvalueIndex(1)).Value.(*wsContext)
	url := L.CheckString(2)

	return libasync.DeferredResult(L, c.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		conn, br, _, err := ws.Dial(ctx, url)
		if err != nil {
			return nil, err
		}

		g := c.luaCtx.ReleasePool().Watch(libpool.NewIOReadWriteCloserResource("websocket conn", conn))
		return func(L *lua.LState) int {
			L.Push(c.class.New(L, &connContext{
				addr:   url,
				conn:   conn,
				br:     br,
				guard:  g,
				luaCtx: c.luaCtx,
			}))
			return 1
		}, nil
	})
}
