package metrics

import (
	lua "github.com/yuin/gopher-lua"

	"github.com/joesonw/distress/pkg/metrics"
)

const metricMetaName = "*METRIC*"

var metricFuncs = map[string]lua.LGFunction{
	"add": lAdd,
}

func lAdd(L *lua.LState) int {
	m := L.CheckUserData(1).Value.(metrics.Metric)
	switch m.Type() {
	case metrics.GaugeMetric:
		m.Add(float64(L.CheckNumber(2)))
	case metrics.CounterMetric:
		m.Add(float64(L.CheckInt(2)))
	case metrics.RateMetric:
		if L.CheckBool(2) {
			m.Add(1)
		} else {
			m.Add(0)
		}
	}
	return 0
}
