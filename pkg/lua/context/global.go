package context

import (
	"sync"

	"github.com/joesonw/lte/pkg/stat"
)

type Global struct {
	reporter stat.Reporter

	uniqueMu  *sync.Mutex
	uniqueMap map[string]interface{}
}

func NewGlobal(reporter stat.Reporter) *Global {
	return &Global{
		reporter:  reporter,
		uniqueMu:  &sync.Mutex{},
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

func (g *Global) Report(stats ...*stat.Stat) {
	g.reporter.Report(stats...)
}
