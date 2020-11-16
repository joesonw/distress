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

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libasync "github.com/joesonw/distress/pkg/lua/lib/async"
	libjson "github.com/joesonw/distress/pkg/lua/lib/json"
	luautil "github.com/joesonw/distress/pkg/lua/util"
	"github.com/joesonw/distress/pkg/metrics"
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
			perfMetric := luautil.NewGlobalUniqueMetric(c.luaCtx.Global(), "*GRPC*_perf", func() metrics.Metric {
				return metrics.Gauge("grpc_perf_us", map[string]string{
					"service": serviceName,
					"method":  methodName,
				})
			}).(metrics.Metric)
			rateMetric := luautil.NewGlobalUniqueMetric(c.luaCtx.Global(), "*GRPC*_rate", func() metrics.Metric {
				return metrics.Rate("grpc_rate", map[string]string{
					"service": serviceName,
					"method":  methodName,
				})
			}).(metrics.Metric)

			bytes, err := libjson.Marshal(reqObj)
			if err != nil {
				return nil, err
			}

			req := dynamic.NewMessage(method.GetInputType())
			if err := req.UnmarshalJSON(bytes); err != nil {
				L.RaiseError(err.Error())
			}

			start := time.Now()
			resMessage, err := c.client.InvokeRpc(ctx, method, req)
			if err != nil {
				rateMetric.Add(0)
				return nil, err
			}
			rateMetric.Add(1)
			perfMetric.Add(float64(time.Since(start).Microseconds()))

			res, err := dynamic.AsDynamicMessage(resMessage)
			if err != nil {
				return nil, err
			}

			bytes, err = res.MarshalJSON()
			if err != nil {
				return nil, err
			}

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
