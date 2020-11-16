package metrics

import (
	"context"
)

type memory struct {
	data *MemoryData
}

type MemoryData struct {
	Counters map[string]struct {
		Count int64
		Tags  map[string]string
	}
	Rate map[string]struct {
		Rate float64
		Tags map[string]string
	}
	Gauge map[string]struct {
		Mean   float64
		Sum    float64
		Min    float64
		Median float64
		Max    float64
		P9999  float64
		P999   float64
		P99    float64
		P95    float64
		P90    float64
		P75    float64
		P50    float64
		Tags   map[string]string
	}
}

func Memory(data *MemoryData) Reporter {
	data.Counters = map[string]struct {
		Count int64
		Tags  map[string]string
	}{}
	data.Rate = map[string]struct {
		Rate float64
		Tags map[string]string
	}{}
	data.Gauge = map[string]struct {
		Mean   float64
		Sum    float64
		Min    float64
		Median float64
		Max    float64
		P9999  float64
		P999   float64
		P99    float64
		P95    float64
		P90    float64
		P75    float64
		P50    float64
		Tags   map[string]string
	}{}
	return &memory{
		data: data,
	}
}

func (memory) isReporter() {}

func (m *memory) Report(ctx context.Context, metrics ...Metric) error {
	for _, metric := range metrics {
		switch data := metric.(type) {
		case *counter:
			{
				var count int64
				for _, v := range data.counts {
					count += v
				}
				m.data.Counters[data.name] = struct {
					Count int64
					Tags  map[string]string
				}{
					Count: count,
					Tags:  data.tags,
				}
			}

		case *rate:
			{
				var truthy float64
				var total float64
				for _, v := range data.values {
					total += 1
					if v {
						truthy += 1
					}
				}
				m.data.Rate[data.name] = struct {
					Rate float64
					Tags map[string]string
				}{
					Rate: truthy / total,
					Tags: data.tags,
				}
			}
		case *gauge:
			m.data.Gauge[data.name] = struct {
				Mean   float64
				Sum    float64
				Min    float64
				Median float64
				Max    float64
				P9999  float64
				P999   float64
				P99    float64
				P95    float64
				P90    float64
				P75    float64
				P50    float64
				Tags   map[string]string
			}{
				Mean:   mustFloat64(data.data.Mean()),
				Sum:    mustFloat64(data.data.Sum()),
				Min:    mustFloat64(data.data.Min()),
				Median: mustFloat64(data.data.Median()),
				Max:    mustFloat64(data.data.Max()),
				P9999:  mustFloat64(data.data.Percentile(99.99)),
				P999:   mustFloat64(data.data.Percentile(99.9)),
				P99:    mustFloat64(data.data.Percentile(99)),
				P95:    mustFloat64(data.data.Percentile(95)),
				P90:    mustFloat64(data.data.Percentile(90)),
				P75:    mustFloat64(data.data.Percentile(75)),
				P50:    mustFloat64(data.data.Percentile(50)),
				Tags:   data.tags,
			}
		}
	}
	return nil
}
