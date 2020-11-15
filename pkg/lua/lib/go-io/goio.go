package goio

import (
	"context"
	"io"
	"io/ioutil"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libasync "github.com/joesonw/distress/pkg/lua/lib/async"
	libbytes "github.com/joesonw/distress/pkg/lua/lib/bytes"
	libpool "github.com/joesonw/distress/pkg/lua/lib/pool"
)

type IO interface {
	GetContext() *luacontext.Context
	GetGuard() *libpool.Guard
}

type Reader interface {
	IO
	io.Reader
}

type Writer interface {
	IO
	io.Writer
}

type Closer interface {
	IO
	io.Closer
}

func Read(L *lua.LState) int {
	ud := L.CheckUserData(1)
	reader, ok := ud.Value.(Reader)
	if !ok {
		L.RaiseError("expected go_io.Reader as UserData for :read()")
	}
	n := L.CheckInt(2)
	return libasync.DeferredResult(L, reader.GetContext().AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		b := make([]byte, n)
		_, err := reader.Read(b)
		if err != nil {
			return nil, err
		}
		return func(L *lua.LState) int {
			L.Push(libbytes.New(L, b))
			return 1
		}, nil
	})
}

func ReadAll(L *lua.LState) int {
	ud := L.CheckUserData(1)
	reader, ok := ud.Value.(Reader)
	if !ok {
		L.RaiseError("expected go_io.Reader as UserData for :read_all()")
	}
	return libasync.DeferredResult(L, reader.GetContext().AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		b, err := ioutil.ReadAll(reader)
		if err != nil {
			return nil, err
		}
		return func(L *lua.LState) int {
			L.Push(libbytes.New(L, b))
			return 1
		}, nil
	})
}

func Write(L *lua.LState) int {
	ud := L.CheckUserData(1)
	writer, ok := ud.Value.(Writer)
	if !ok {
		L.RaiseError("expected go_io.Writer as UserData for :write()")
	}
	bytes := libbytes.Check(L, 2)
	return libasync.Deferred(L, writer.GetContext().AsyncPool(), func(ctx context.Context) error {
		_, err := writer.Write(bytes)
		return err
	})
}

func Close(L *lua.LState) int {
	ud := L.CheckUserData(1)
	closer, ok := ud.Value.(Closer)
	if !ok {
		L.RaiseError("expected go_io.Closer as UserData for :close()")
	}
	return libasync.Deferred(L, closer.GetContext().AsyncPool(), func(ctx context.Context) error {
		if g := closer.GetGuard(); g != nil {
			g.Done()
		}
		return closer.Close()
	})
}
