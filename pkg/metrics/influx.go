package metrics

import (
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	influxdb2api "github.com/influxdata/influxdb-client-go/v2/api"
	"go.uber.org/multierr"
)

type influx struct {
	name     string
	writeAPI influxdb2api.WriteAPI
	metrics  []Metric
	errs     []error
	ticker   *time.Ticker
}

func Influx(writeAPI influxdb2api.WriteAPI, interval time.Duration, name string) Reporter {
	r := &influx{
		writeAPI: writeAPI,
		ticker:   time.NewTicker(interval),
	}
	r.startTick()
	return r
}

func (influx) isReporter() {}

func (r *influx) Collect(metrics ...Metric) {
	r.metrics = append(r.metrics, metrics...)
}

func (r *influx) Finish() error {
	r.ticker.Stop()
	return multierr.Combine(r.errs...)
}

func (r *influx) startTick() {
	ch := r.writeAPI.Errors()
	go func() {
		select {
		case err := <-ch:
			r.errs = append(r.errs, err)
		case <-r.ticker.C:
			r.tick()
		}
	}()
}

func (r *influx) tick() {

	for _, metric := range r.metrics {
		switch data := metric.(type) {
		case *counter:
			{
				r.writeAPI.WritePoint(influxdb2.NewPoint(
					data.name,
					data.tags,
					map[string]interface{}{
						"count": data.value,
					},
					time.Now()).AddTag("job", r.name))
			}
		case *gauge:
			{
				r.writeAPI.WritePoint(influxdb2.NewPoint(
					data.name,
					data.tags,
					map[string]interface{}{
						"sum":   mustFloat64(data.data.Sum()),
						"count": data.data.Len(),
					},
					time.Now()).AddTag("job", r.name))

			}
		case *rate:
			{
				r.writeAPI.WritePoint(influxdb2.NewPoint(
					data.name,
					data.tags,
					map[string]interface{}{
						"value": data.Value(),
					},
					time.Now()).AddTag("job", r.name))
			}
		}
	}
	r.writeAPI.Flush()
}
