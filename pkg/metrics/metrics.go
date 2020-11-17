package metrics

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/montanaflynn/stats"
)

type Type uint

func (t Type) String() string {
	return typeNames[t]
}

var typeNames = map[Type]string{
	CounterMetric: "COUNTER",
	GaugeMetric:   "GAUGE",
	RateMetric:    "RATE",
}

const (
	CounterMetric Type = iota
	GaugeMetric
	RateMetric
)

type Reporter interface {
	isReporter()
	Finish() error
	Collect(metrics ...Metric)
}

type Metric interface {
	Add(float64)
	Type() Type
}

type counter struct {
	name  string
	value int64
	tags  map[string]string
}

func (c *counter) Add(v float64) {
	atomic.AddInt64(&c.value, int64(v))
}

func (c *counter) Value() int64 {
	return c.value
}

func (c *counter) Type() Type {
	return CounterMetric
}

func Counter(name string, tags map[string]string) Metric {
	return &counter{
		name: name,
		tags: tags,
	}
}

type gauge struct {
	mu   *sync.Mutex
	name string
	tags map[string]string
	data stats.Float64Data
}

func (c *gauge) Add(v float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = append(c.data, v)
}

func (c *gauge) Value() stats.Float64Data {
	return c.data[:]
}

func (c *gauge) Type() Type {
	return GaugeMetric
}

func Gauge(name string, tags map[string]string) Metric {
	return &gauge{
		name: name,
		tags: tags,
		mu:   &sync.Mutex{},
	}
}

type rate struct {
	mu     *sync.Mutex
	values []bool
	name   string
	tags   map[string]string
}

func (c *rate) Add(v float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values = append(c.values, v == 1)
}

func (c *rate) Value() float64 {
	var truthy float64
	var total float64
	for _, v := range c.values {
		total += 1
		if v {
			truthy += 1
		}
	}
	return truthy / total
}

func (c *rate) Type() Type {
	return RateMetric
}

func Rate(name string, tags map[string]string) Metric {
	return &rate{
		mu:   &sync.Mutex{},
		name: name,
		tags: tags,
	}
}

func sprintFloat64(v float64, _ error) string {
	return strconv.FormatFloat(v, 'f', 1, 64)
}

func sprintInt(i int) string {
	return strconv.FormatInt(int64(i), 10)
}

func sprintInt64(i int64) string {
	return strconv.FormatInt(i, 10)
}

func sprintTags(tags map[string]string) string {
	var kvs []string
	for k, v := range tags {
		kvs = append(kvs, fmt.Sprintf("%s=%s", k, v))
	}
	return strings.Join(kvs, "\n")
}

func mustFloat64(v float64, _ error) float64 {
	return v
}
