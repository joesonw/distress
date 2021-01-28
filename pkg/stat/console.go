package stat

import (
	"fmt"
	"strings"
)

func Console() Reporter {
	return console{}
}

type console struct{}

func (console) Report(stats ...*Stat) {
	for _, stat := range stats {
		var tags []string
		var fields []string
		for k, v := range stat.Tags {
			tags = append(tags, fmt.Sprintf(",%s=%s", k, v))
		}
		for k, v := range stat.Fields {
			fields = append(fields, fmt.Sprintf("%s=%f", k, v))
		}
		fmt.Printf("%s%s %s %d", stat.Name, strings.Join(tags, ""), strings.Join(fields, ","), stat.Timestamp.UnixNano())
		println()
	}
}

func (console) Finish() error { return nil }
