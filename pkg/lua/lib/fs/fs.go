package fs

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/spf13/afero"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libasync "github.com/joesonw/distress/pkg/lua/lib/async"
	libbytes "github.com/joesonw/distress/pkg/lua/lib/bytes"
	goclass "github.com/joesonw/distress/pkg/lua/lib/go-class"
	libpool "github.com/joesonw/distress/pkg/lua/lib/pool"
	libtime "github.com/joesonw/distress/pkg/lua/lib/time"
)

const moduleName = "fs"

var fsFuncs = map[string]lua.LGFunction{
	"open":       fsOpen,
	"create":     fsCreate,
	"remove":     fsRemove,
	"remove_all": fsRemoveAll,
	"list":       fsList,
	"stat":       fsStat,
	"mkdir_all":  fsMkdirAll,
	"exist":      fsExist,
	"read_all":   fsReadAll,
	"write":      fsWrite,
}

type fsContext struct {
	fs        afero.Fs
	fileClass *goclass.Class
	luaCtx    *luacontext.Context
}

func upFS(L *lua.LState) *fsContext {
	return L.CheckUserData(lua.UpvalueIndex(1)).Value.(*fsContext)
}

func Open(L *lua.LState, luaCtx *luacontext.Context, fs afero.Fs) {
	mod := L.RegisterModule(moduleName, map[string]lua.LGFunction{}).(*lua.LTable)
	ctx := &fsContext{
		fs:     fs,
		luaCtx: luaCtx,
	}
	uv := L.NewUserData()
	uv.Value = ctx
	for name, f := range fsFuncs {
		mod.RawSetString(name, L.NewClosure(f, uv))
	}

	ctx.fileClass = goclass.New(L, fileMetaName, fileFuncs)
}

func fsOpen(L *lua.LState) int {
	fs := upFS(L)
	name := L.CheckString(2)
	return libasync.DeferredResult(L, fs.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		f, err := fs.fs.OpenFile(name, os.O_CREATE|os.O_RDWR, 0777)
		if err != nil {
			return nil, err
		}
		return func(L *lua.LState) int {
			g := fs.luaCtx.ReleasePool().Watch(libpool.NewOSFileResource(f))
			L.Push(fs.fileClass.New(L, &file{
				name:  name,
				File:  f,
				ctx:   fs,
				guard: g,
			}))
			return 1
		}, nil
	})
}

func fsCreate(L *lua.LState) int {
	fs := upFS(L)
	name := L.CheckString(2)
	return libasync.DeferredResult(L, fs.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		f, err := fs.fs.Create(name)
		if err != nil {
			return nil, err
		}
		return func(L *lua.LState) int {
			g := fs.luaCtx.ReleasePool().Watch(libpool.NewOSFileResource(f))
			L.Push(fs.fileClass.New(L, &file{
				name:  name,
				File:  f,
				ctx:   fs,
				guard: g,
			}))
			return 1
		}, nil
	})
}

func fsRemove(L *lua.LState) int {
	fs := upFS(L)
	name := L.CheckString(2)
	return libasync.Deferred(L, fs.luaCtx.AsyncPool(), func(ctx context.Context) error {
		return fs.fs.Remove(name)
	})
}

func fsRemoveAll(L *lua.LState) int {
	fs := upFS(L)
	name := L.CheckString(2)
	return libasync.Deferred(L, fs.luaCtx.AsyncPool(), func(ctx context.Context) error {
		return fs.fs.RemoveAll(name)
	})
}

func fsList(L *lua.LState) int {
	fs := upFS(L)
	name := L.CheckString(2)
	return libasync.DeferredResult(L, fs.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		list, err := afero.ReadDir(fs.fs, name)
		if err != nil {
			return nil, err
		}

		return func(L *lua.LState) int {
			l := L.NewTable()
			for i := range list {
				info := list[i]
				f := L.NewTable()
				f.RawSetString("name", lua.LString(info.Name()))
				f.RawSetString("size", lua.LNumber(info.Size()))
				f.RawSetString("dir", lua.LBool(info.IsDir()))
				f.RawSetString("modtime", libtime.New(L, info.ModTime()))
				l.Append(f)
			}
			L.Push(l)
			return 1
		}, nil
	})
}

func fsStat(L *lua.LState) int {
	fs := upFS(L)
	name := L.CheckString(2)
	return libasync.DeferredResult(L, fs.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		info, err := fs.fs.Stat(name)
		if err != nil {
			return nil, err
		}

		return func(L *lua.LState) int {
			f := L.NewTable()
			f.RawSetString("name", lua.LString(info.Name()))
			f.RawSetString("size", lua.LNumber(info.Size()))
			f.RawSetString("dir", lua.LBool(info.IsDir()))
			f.RawSetString("modtime", libtime.New(L, info.ModTime()))
			L.Push(f)
			return 1
		}, nil
	})
}

func fsMkdirAll(L *lua.LState) int {
	fs := upFS(L)
	name := L.CheckString(2)
	return libasync.Deferred(L, fs.luaCtx.AsyncPool(), func(ctx context.Context) error {
		return fs.fs.Mkdir(name, 0777)
	})
}

func fsExist(L *lua.LState) int {
	fs := upFS(L)
	name := L.CheckString(2)
	return libasync.DeferredResult(L, fs.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		ok, err := afero.Exists(fs.fs, name)
		if err != nil {
			return nil, err
		}
		return func(L *lua.LState) int {
			L.Push(lua.LBool(ok))
			return 1
		}, nil
	})
}

func fsReadAll(L *lua.LState) int {
	fs := upFS(L)
	name := L.CheckString(2)
	return libasync.DeferredResult(L, fs.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		f, err := fs.fs.OpenFile(name, os.O_RDONLY, 0777)
		if err != nil {
			return nil, err
		}
		b, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}
		return func(L *lua.LState) int {
			L.Push(libbytes.New(L, b))
			return 1
		}, nil
	})
}

func fsWrite(L *lua.LState) int {
	fs := upFS(L)
	name := L.CheckString(2)
	contents := libbytes.Check(L, 3)
	return libasync.Deferred(L, fs.luaCtx.AsyncPool(), func(ctx context.Context) error {
		f, err := fs.fs.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0777)
		if err != nil {
			return err
		}
		_, err = f.Write(contents)
		return err
	})
}
