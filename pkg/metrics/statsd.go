package metrics

import (
	"fmt"
	"time"

	gostatsd "github.com/DataDog/datadog-go/statsd"
	"go.uber.org/multierr"
)

type statsd struct {
	name    string
	metrics []Metric
	client  *gostatsd.Client
	errs    []error
	ticker  *time.Ticker
}

func Statsd(client *gostatsd.Client, interval time.Duration, name string) Reporter {
	r := &statsd{
		name:   name,
		client: client,
		ticker: time.NewTicker(interval),
	}
	r.startTick()
	return r
}

func (statsd) isReporter() {}

func (r *statsd) Collect(metrics ...Metric) {
	r.metrics = append(r.metrics, metrics...)
}

func (r *statsd) Finish() error {
	r.ticker.Stop()
	return multierr.Combine(r.errs...)
}

func (r *statsd) startTick() {
	go func() {
		for range r.ticker.C {
			r.tick()
		}
	}()
}

func (r *statsd) tick() {

	for _, metric := range r.metrics {
		switch data := metric.(type) {
		case *counter:
			if err := r.client.Count(data.name, data.Value(), append(makeStatsdTags(data.tags), "job:"+r.name), 1); err != nil {
				r.errs = append(r.errs, err)
			}
		case *gauge:
			if err := r.client.Distribution(data.name, mustFloat64(data.data.Mean()), append(makeStatsdTags(data.tags), "job:"+r.name), 1); err != nil {
				r.errs = append(r.errs, err)
			}
		case *rate:
			if err := r.client.Gauge(data.name, data.Value(), append(makeStatsdTags(data.tags), "job:"+r.name), 1); err != nil {
				r.errs = append(r.errs, err)
			}
		}
	}
	if err := r.client.Flush(); err != nil {
		r.errs = append(r.errs, err)
	}
}

func makeStatsdTags(m map[string]string) []string {
	var tags []string
	for k, v := range m {
		tags = append(tags, fmt.Sprintf("%s:%s", k, v))
	}
	return tags
}
