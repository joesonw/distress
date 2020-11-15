package context

import (
	"context"
	"sync"

	"github.com/joesonw/distress/pkg/metrics"
)

type Global struct {
	uniqueMu  *sync.Mutex
	uniqueMap map[string]interface{}

	metricsMu *sync.Mutex
	metrics   []metrics.Metric
}

func NewGlobal() *Global {
	return &Global{
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
}

func (g *Global) ReportMetrics(ctx context.Context, r metrics.Reporter) error {
	return r.Report(ctx, g.metrics...)
}
