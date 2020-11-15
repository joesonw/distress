package websocket

import (
	"bufio"
	"context"
	"net"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libasync "github.com/joesonw/distress/pkg/lua/lib/async"
	libbytes "github.com/joesonw/distress/pkg/lua/lib/bytes"
	libpool "github.com/joesonw/distress/pkg/lua/lib/pool"
)

const connMetaName = "*WEBSOCKET*CONN*"

type connContext struct {
	messages []wsutil.Message
	conn     net.Conn
	br       *bufio.Reader
	guard    *libpool.Guard
	luaCtx   *luacontext.Context
}

var connFuncs = map[string]lua.LGFunction{
	"read":  connRead,
	"write": connWrite,
	"close": connClose,
}

func connRead(L *lua.LState) int {
	c := L.CheckUserData(1).Value.(*connContext)
	return libasync.DeferredResult(L, c.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		if c.br != nil {
			messages, err := wsutil.ReadServerMessage(c.br, nil)
			ws.PutReader(c.br)
			c.br = nil
			if err != nil {
				return nil, err
			}
			for i := range messages {
				if messages[i].OpCode.IsData() {
					c.messages = append(c.messages, messages[i])
				}
			}
		}

		if len(c.messages) > 0 {
			m := c.messages[len(c.messages)-1]
			c.messages = c.messages[:len(c.messages)-1]
			return func(L *lua.LState) int {
				L.Push(libbytes.New(L, m.Payload))
				return 1
			}, nil
		}

		b, err := wsutil.ReadServerText(c.conn)
		if err != nil {
			return nil, err
		}
		return func(L *lua.LState) int {
			L.Push(libbytes.New(L, b))
			return 1
		}, nil
	})
}

func connWrite(L *lua.LState) int {
	c := L.CheckUserData(1).Value.(*connContext)
	bytes := libbytes.Check(L, 2)
	return libasync.Deferred(L, c.luaCtx.AsyncPool(), func(ctx context.Context) error {
		return wsutil.WriteClientText(c.conn, bytes)
	})
}

func connClose(L *lua.LState) int {
	c := L.CheckUserData(1).Value.(*connContext)
	return libasync.Deferred(L, c.luaCtx.AsyncPool(), func(ctx context.Context) error {
		if c.br != nil {
			ws.PutReader(c.br)
		}
		c.guard.Done()
		return c.conn.Close()
	})
}
