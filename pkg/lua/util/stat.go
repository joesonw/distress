package util

import (
	luacontext "github.com/joesonw/lte/pkg/lua/context"
	"github.com/joesonw/lte/pkg/stat"
)

func ReportContextStat(ctx *luacontext.Context, stats ...*stat.Stat) {
	scopeName := ctx.ScopeName()
	tags := ctx.Tags()
	for _, s := range stats {
		s.Tag("scope", scopeName)
		for k, v := range tags {
			s.Tag("scope:"+k, v)
		}
	}
	ctx.Global().Report(stats...)
}
