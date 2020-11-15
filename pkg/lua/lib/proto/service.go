package proto

import (
	protodesc "github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
)

type serviceContext struct {
	desc   *protodesc.ServiceDescriptor
	luaCtx *luacontext.Context
}

func newService(L *lua.LState, luaCtx *luacontext.Context, desc *protodesc.ServiceDescriptor) lua.LValue {
	ud := L.NewUserData()
	ud.Value = &serviceContext{
		desc:   desc,
		luaCtx: luaCtx,
	}

	index := L.NewTable()
	index.RawSetString("new", L.NewClosure(serviceNew, ud))

	meta := L.NewTable()
	meta.RawSetString("__index", index)
	L.SetMetatable(ud, meta)
	return ud
}

func serviceNew(L *lua.LState) int {
	c := L.CheckUserData(lua.UpvalueIndex(1)).Value.(*serviceContext)
	ud := L.CheckUserData(2)
	gc, ok := ud.Value.(*grpcClientConnContext)
	if !ok {
		L.RaiseError("expected a grpc connection")
	}

	client := grpcdynamic.NewStub(gc.cc)
	L.Push(newServiceClient(L, c.luaCtx, c.desc, client))
	return 1
}
