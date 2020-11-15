package metrics

import (
	"context"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"go.uber.org/multierr"
)

type influx struct {
	client      influxdb2.Client
	org, bucket string
}

func Influx(client influxdb2.Client, org, bucket string) Reporter {
	return &influx{
		client: client,
		org:    org,
		bucket: bucket,
	}
}

func (influx) isReporter() {}

func (r *influx) Report(ctx context.Context, metrics ...Metric) error {
	write := r.client.WriteAPI(r.org, r.bucket)
	errCh := write.Errors()
	var errs []error
	go func() {
		for err := range errCh {
			errs = append(errs, err)
		}
	}()
	for _, metric := range metrics {
		switch data := metric.(type) {
		case *counter:
			{
				var count int64
				for i := range data.counts {
					count += data.counts[i]
					write.WritePoint(influxdb2.NewPoint(
						data.name,
						data.tags,
						map[string]interface{}{
							"count": count,
						},
						data.timestamp[i]))
				}
			}
		case *gauge:
			{
				var sum float64
				for i := range data.data {
					sum += data.data[i]
					write.WritePoint(influxdb2.NewPoint(
						data.name,
						data.tags,
						map[string]interface{}{
							"sum": sum,
						},
						data.timestamp[i]))
				}

			}
		case *rate:
			{
				for i := range data.values {
					write.WritePoint(influxdb2.NewPoint(
						data.name,
						data.tags,
						map[string]interface{}{
							"value": data.values[i],
						},
						data.timestamp[i]))
				}
			}
		}
	}
	write.Flush()
	return multierr.Combine(errs...)
}
