package proto

import (
	"os"

	"github.com/jhump/protoreflect/desc/protoparse"
	"github.com/spf13/afero"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	goclass "github.com/joesonw/lte/pkg/lua/lib/go-class"
)

const moduleName = "proto"

type protoContext struct {
	fs           afero.Fs
	messageClass *goclass.Class
	luaCtx       *luacontext.Context
}

func Open(L *lua.LState, luaCtx *luacontext.Context, fs afero.Fs) {
	mod := L.RegisterModule(moduleName, map[string]lua.LGFunction{}).(*lua.LTable)
	ud := L.NewUserData()
	ud.Value = &protoContext{
		fs:           fs,
		messageClass: goclass.New(L, messageMetaName, messageFuncs),
		luaCtx:       luaCtx,
	}
	mod.RawSetString("load", L.NewClosure(protoLoad, ud))
	mod.RawSetString("dial", L.NewClosure(protoDial, ud))
}

func protoLoad(L *lua.LState) int {
	c := L.CheckUserData(lua.UpvalueIndex(1)).Value.(*protoContext)
	files := map[string]string{}

	err := afero.Walk(c.fs, "", func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			bytes, err := afero.ReadFile(c.fs, path)
			if err != nil {
				return err
			}
			files[path] = string(bytes)
		}
		return err
	})
	if err != nil {
		L.RaiseError(err.Error())
	}

	fileAccessor := protoparse.FileContentsFromMap(files)
	parser := protoparse.Parser{
		Accessor: fileAccessor,
		WarningReporter: func(err protoparse.ErrorWithPos) {
			c.luaCtx.Logger().With(zap.String("module", "lua-proto")).Warn(err.Error())
		},
	}

	var target []string
	for i := 2; i < L.GetTop()+1; i++ {
		target = append(target, L.CheckString(i))
	}

	fileDescs, err := parser.ParseFiles(target...)
	if err != nil {
		L.RaiseError(err.Error())
	}

	services := L.NewTable()
	messages := L.NewTable()
	for _, fileDesc := range fileDescs {
		for _, desc := range fileDesc.GetMessageTypes() {
			messages.RawSetString(desc.GetFullyQualifiedName(), c.messageClass.New(L, &messageContext{desc: desc, luaCtx: c.luaCtx}))
		}
		for _, desc := range fileDesc.GetServices() {
			services.RawSetString(desc.GetFullyQualifiedName(), newService(L, c.luaCtx, desc))
		}
	}

	L.Push(messages)
	L.Push(services)
	return 2
}
