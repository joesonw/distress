package util

import (
	luacontext "github.com/joesonw/lte/pkg/lua/context"
	"github.com/joesonw/lte/pkg/metrics"
)

func NewGlobalUniqueMetric(global *luacontext.Global, name string, f func() metrics.Metric) metrics.Metric {
	return global.Unique(name, func() interface{} {
		m := f()
		global.RegisterMetric(m)
		return m
	}).(metrics.Metric)
}
