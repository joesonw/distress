package time_test

import (
	"testing"
	"time"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libtime "github.com/joesonw/lte/pkg/lua/lib/time"
	test_util "github.com/joesonw/lte/pkg/lua/test-util"
)

func Test(t *testing.T) {
	test_util.Run(t, func(t *testing.T) *test_util.Test {
		return test_util.New(
			"time",
			`
					assert(t:string() == "2006-01-02 15:04:05 +0000 UTC", "time:string")
				`,
		).Before(func(t *testing.T, L *lua.LState, luaCtx *luacontext.Context) {
			v, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
			L.SetGlobal("t", libtime.New(L, v))
		})
	})
}
