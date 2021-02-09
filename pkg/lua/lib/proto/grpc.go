package proto

import (
	"context"
	"fmt"
	"time"

	protodesc "github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	"github.com/jhump/protoreflect/dynamic/grpcdynamic"
	lua "github.com/yuin/gopher-lua"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libasync "github.com/joesonw/lte/pkg/lua/lib/async"
	libjson "github.com/joesonw/lte/pkg/lua/lib/json"
	luautil "github.com/joesonw/lte/pkg/lua/util"
	"github.com/joesonw/lte/pkg/stat"
)

type grpcClientConnContext struct {
	cc     *grpc.ClientConn
	luaCtx *luacontext.Context
}

func protoDial(L *lua.LState) int {
	c := L.CheckUserData(lua.UpvalueIndex(1)).Value.(*protoContext)
	addr := L.CheckString(2)

	optionalOpts := L.Get(3)
	opts, ok := optionalOpts.(*lua.LTable)
	var dialOpts []grpc.DialOption

	if ok {
		if val := opts.RawGetString("insecure"); val == lua.LTrue {
			dialOpts = append(dialOpts, grpc.WithInsecure())
		}
	}

	return libasync.DeferredResult(L, c.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		cc, err := grpc.Dial(addr, dialOpts...)
		if err != nil {
			return nil, err
		}

		return func(L *lua.LState) int {
			ud := L.NewUserData()
			ud.Value = &grpcClientConnContext{
				cc:     cc,
				luaCtx: c.luaCtx,
			}
			L.Push(ud)
			return 1
		}, nil
	})
}

type serviceClientContext struct {
	client grpcdynamic.Stub
	desc   *protodesc.ServiceDescriptor
	luaCtx *luacontext.Context
}

func newServiceClient(L *lua.LState, luaCtx *luacontext.Context, desc *protodesc.ServiceDescriptor, client grpcdynamic.Stub) lua.LValue {
	ud := L.NewUserData()
	ud.Value = &serviceClientContext{
		client: client,
		desc:   desc,
		luaCtx: luaCtx,
	}

	index := L.NewTable()
	index.RawSetString("new", L.NewClosure(serviceNew, ud))

	meta := L.NewTable()
	meta.RawSetString("__index", L.NewClosure(serviceClientCall, ud))
	L.SetMetatable(ud, meta)
	return ud
}

func serviceClientCall(L *lua.LState) int {
	c := L.CheckUserData(lua.UpvalueIndex(1)).Value.(*serviceClientContext)
	name := L.CheckString(2)
	L.Push(L.NewClosure(func(L *lua.LState) int {
		reqObj := L.CheckTable(2)
		return libasync.DeferredResult(L, c.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
			method := c.desc.FindMethodByName(name)
			if method == nil {
				return nil, fmt.Errorf("method \"%s\" not found", name)
			}
			serviceName := c.desc.GetFullyQualifiedName()
			methodName := method.GetFullyQualifiedName()

			bytes, err := libjson.Marshal(reqObj)
			if err != nil {
				return nil, err
			}

			s := stat.New("grpc").Tag("service", serviceName).Tag("method", methodName)
			defer luautil.ReportContextStat(c.luaCtx, s)

			req := dynamic.NewMessage(method.GetInputType())
			if err := req.UnmarshalJSON(bytes); err != nil {
				L.RaiseError(err.Error())
			}

			start := time.Now()
			resMessage, err := c.client.InvokeRpc(ctx, method, req)
			if err != nil {
				code := status.Code(err)
				s.Tag("code", code.String()).Int64Field("success", 0)
				return nil, err
			}

			s.Int64Field("duration_ns", time.Since(start).Nanoseconds())
			res, err := dynamic.AsDynamicMessage(resMessage)
			if err != nil {
				s.Int64Field("success", 0)
				return nil, err
			}

			bytes, err = res.MarshalJSON()
			if err != nil {
				s.Int64Field("success", 0)
				return nil, err
			}
			s.IntField("response_size", len(bytes))

			s.Int64Field("success", 1).IntField("response_size", len(bytes))

			return func(L *lua.LState) int {
				val, err := libjson.Unmarshal(L, bytes)
				if err != nil {
					L.RaiseError(err.Error())
				}

				L.Push(val)
				return 1
			}, nil
		})
	}))

	return 1
}
