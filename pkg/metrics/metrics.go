package metrics

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

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
	Report(ctx context.Context, metrics ...Metric) error
}

type Metric interface {
	Add(float64)
	Type() Type
}

type counter struct {
	mu        *sync.Mutex
	name      string
	counts    []int64
	tags      map[string]string
	timestamp []time.Time
}

func (c *counter) Add(v float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.counts = append(c.counts, int64(v))
	c.timestamp = append(c.timestamp, time.Now())
}

func (c *counter) Type() Type {
	return CounterMetric
}

func Counter(name string, tags map[string]string) Metric {
	return &counter{
		mu:   &sync.Mutex{},
		name: name,
		tags: tags,
	}
}

type gauge struct {
	mu        *sync.Mutex
	name      string
	tags      map[string]string
	data      stats.Float64Data
	timestamp []time.Time
}

func (c *gauge) Add(v float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data = append(c.data, v)
	c.timestamp = append(c.timestamp, time.Now())
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
	mu        *sync.Mutex
	values    []bool
	name      string
	tags      map[string]string
	timestamp []time.Time
}

func (c *rate) Add(v float64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values = append(c.values, v == 1)
	c.timestamp = append(c.timestamp, time.Now())
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
	return strings.Join(kvs, ",")
}

func mustFloat64(v float64, _ error) float64 {
	return v
}
