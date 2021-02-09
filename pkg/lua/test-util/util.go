package test_util

import (
	"math/rand"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libbase "github.com/joesonw/lte/pkg/lua/lib/base"
	libbytes "github.com/joesonw/lte/pkg/lua/lib/bytes"
	libjson "github.com/joesonw/lte/pkg/lua/lib/json"
	libpool "github.com/joesonw/lte/pkg/lua/lib/pool"
	libtime "github.com/joesonw/lte/pkg/lua/lib/time"
	"github.com/joesonw/lte/pkg/stat"
)

type Before func(t *testing.T, L *lua.LState, luaCtx *luacontext.Context)
type After func(t *testing.T, L *lua.LState)

type Testable func(*testing.T) *Test

type Test struct {
	name   string
	script string
	before Before
	after  After
}

func (t *Test) Before(before Before) *Test {
	t.before = before
	return t
}

func (t *Test) After(after After) *Test {
	t.after = after
	return t
}

func New(name, script string) *Test {
	return &Test{
		name:   name,
		script: script,
	}
}

func Run(t *testing.T, tests ...Testable) {
	rand.Seed(time.Now().UnixNano())
	for _, fn := range tests {
		test := fn(t)
		t.Run(test.name, func(t *testing.T) {
			L := lua.NewState(lua.Options{
				CallStackSize:       0,
				RegistrySize:        0,
				RegistryMaxSize:     0,
				RegistryGrowStep:    0,
				SkipOpenLibs:        true,
				IncludeGoStackTrace: false,
				MinimizeStackMemory: false,
			})
			defer L.Close()

			logger, _ := zap.NewDevelopment()
			defer logger.Sync() //nolint:errcheck

			asyncPool := libpool.NewAsync(logger, 4, 0, 16)
			asyncPool.Start()
			defer asyncPool.Stop()

			releasePool := libpool.NewRelease(logger)
			defer releasePool.Clean()

			reporter := stat.Console()
			luaCtx := luacontext.New(L, luacontext.NewGlobal(reporter), releasePool, asyncPool, logger)

			lua.OpenBase(L)
			lua.OpenPackage(L)
			lua.OpenMath(L)
			lua.OpenString(L)
			lua.OpenTable(L)
			lua.OpenOs(L)

			libbase.Open(L, luaCtx, afero.NewMemMapFs())
			libjson.Open(L, luaCtx)
			libbytes.Open(L, luaCtx)

			libtime.Open(L, luaCtx)

			if before := test.before; before != nil {
				before(t, L, luaCtx)
			}

			err := L.DoString(test.script)
			assert.Nil(t, err)

			if after := test.after; after != nil {
				after(t, L)
			}
		})
	}
}
