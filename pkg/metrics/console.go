package metrics

import (
	"bytes"
	"fmt"

	"github.com/olekukonko/tablewriter"
)

type console struct {
	metrics []Metric
}

func Console() Reporter {
	return &console{}
}

func (console) isReporter() {}

func (c *console) Finish() error {
	buf := bytes.NewBuffer(nil)
	counterTable := tablewriter.NewWriter(buf)
	counterTable.SetRowLine(true)
	counterTable.SetHeader([]string{"Name", "Tags", "Count"})

	rateTable := tablewriter.NewWriter(buf)
	rateTable.SetRowLine(true)
	rateTable.SetHeader([]string{"Name", "Tags", "True%"})

	gaugeTable := tablewriter.NewWriter(buf)
	gaugeTable.SetRowLine(true)
	gaugeTable.SetHeader([]string{"Name", "Tags", "Count", "Avg", "Total", "Min", "Med", "Max", "99.99%", "99.9%", "99%", "95%", "90%", "75%", "50%"})

	for _, metric := range c.metrics {
		switch data := metric.(type) {
		case *counter:
			counterTable.Append([]string{
				data.name,
				sprintTags(data.tags),
				sprintInt64(data.value),
			})

		case *rate:
			rateTable.Append([]string{
				data.name,
				sprintTags(data.tags),
				sprintFloat64(data.Value()*100, nil) + "%",
			})
		case *gauge:
			gaugeTable.Append([]string{
				data.name,
				sprintTags(data.tags),
				sprintInt(data.data.Len()),
				sprintFloat64(data.data.Mean()),
				sprintFloat64(data.data.Sum()),
				sprintFloat64(data.data.Min()),
				sprintFloat64(data.data.Median()),
				sprintFloat64(data.data.Max()),
				sprintFloat64(data.data.Percentile(99.99)),
				sprintFloat64(data.data.Percentile(99.9)),
				sprintFloat64(data.data.Percentile(99)),
				sprintFloat64(data.data.Percentile(95)),
				sprintFloat64(data.data.Percentile(90)),
				sprintFloat64(data.data.Percentile(75)),
				sprintFloat64(data.data.Percentile(50)),
			})
		}
	}

	buf.WriteString("Counters:\n")
	counterTable.Render()

	buf.WriteString("Gauges:\n")
	gaugeTable.Render()

	buf.WriteString("Rates:\n")
	rateTable.Render()

	fmt.Println(buf.String())
	return nil
}

func (c *console) Collect(metrics ...Metric) {
	c.metrics = append(c.metrics, metrics...)
}
