package metrics

import (
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	goclass "github.com/joesonw/distress/pkg/lua/lib/go-class"
	"github.com/joesonw/distress/pkg/metrics"
)

const moduleName = "metrics"

type modContext struct {
	luaCtx *luacontext.Context
	class  *goclass.Class
}

func Open(L *lua.LState, luaCtx *luacontext.Context) {
	mod := L.RegisterModule(moduleName, map[string]lua.LGFunction{}).(*lua.LTable)
	ud := L.NewUserData()
	ud.Value = &modContext{
		luaCtx: luaCtx,
		class:  goclass.New(L, metricMetaName, metricFuncs),
	}
	L.SetFuncs(mod, map[string]lua.LGFunction{
		"gauge": makeMetricsFunction(func(name string, tags map[string]string) metrics.Metric {
			return metrics.Gauge(name, tags)
		}),
		"rate": makeMetricsFunction(func(name string, tags map[string]string) metrics.Metric {
			return metrics.Rate(name, tags)
		}),
		"counter": makeMetricsFunction(func(name string, tags map[string]string) metrics.Metric {
			return metrics.Counter(name, tags)
		}),
	}, ud)
}

func makeMetricsFunction(f func(name string, tags map[string]string) metrics.Metric) lua.LGFunction {
	return func(L *lua.LState) int {
		c := L.CheckUserData(1).Value.(*modContext)
		name := L.CheckString(2)
		metric := c.luaCtx.Global().Unique("*METRIC*-"+name, func() interface{} {
			tags := map[string]string{}
			if v := L.Get(3); v.Type() == lua.LTTable {
				v.(*lua.LTable).ForEach(func(k, v lua.LValue) {
					tags[k.String()] = v.String()
				})
			}
			scopeTags := c.luaCtx.Tags()
			for k := range scopeTags {
				tags[k] = scopeTags[k]
			}
			metric := f(name, tags)
			c.luaCtx.Global().RegisterMetric(metric)
			return metric
		})
		L.Push(c.class.New(L, metric))
		return 1
	}
}
