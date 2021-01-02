package context

import (
	"sync"

	"github.com/joesonw/lte/pkg/metrics"
)

type Global struct {
	reporter metrics.Reporter

	uniqueMu  *sync.Mutex
	uniqueMap map[string]interface{}
	metricsMu *sync.Mutex
	metrics   []metrics.Metric
}

func NewGlobal(reporter metrics.Reporter) *Global {
	return &Global{
		reporter:  reporter,
		uniqueMu:  &sync.Mutex{},
		metricsMu: &sync.Mutex{},
		uniqueMap: map[string]interface{}{},
	}
}

func (g *Global) Unique(name string, do func() interface{}) interface{} {
	g.uniqueMu.Lock()
	defer g.uniqueMu.Unlock()
	if in := g.uniqueMap[name]; in != nil {
		return in
	}
	in := do()
	g.uniqueMap[name] = in
	return in
}

func (g *Global) RegisterMetric(m metrics.Metric) {
	g.metricsMu.Lock()
	defer g.metricsMu.Unlock()
	g.metrics = append(g.metrics, m)
	g.reporter.Collect(m)
}
