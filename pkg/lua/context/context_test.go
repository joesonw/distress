package context_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest"
	"go.uber.org/zap/zaptest/observer"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
)

func TestContext(t *testing.T) {
	L := lua.NewState()

	logger := zaptest.NewLogger(t)
	core, observedLogs := observer.New(logger.Core())
	logger = zaptest.NewLogger(t, zaptest.WrapOptions(zap.WrapCore(func(zapcore.Core) zapcore.Core {
		return core
	})))

	ctx := luacontext.New(L, nil, nil, nil, logger)

	assert.Equal(t, "", ctx.Scope())
	assert.Equal(t, 0, len(ctx.Tags()))
	ctx.Logger().Info("init")

	ctx.Enter("out", "key", "value")
	assert.Equal(t, "value", ctx.Tags()["key"])
	ctx.Logger().Info("out")
	assert.Equal(t, "out(key=value)", ctx.Scope())

	L.SetGlobal("run", L.NewFunction(func(L *lua.LState) int {
		ctx.Logger().Info("inside")
		assert.Equal(t, "out.test.nested(key=value,key2=value2)", ctx.Scope())
		assert.Equal(t, "value", ctx.Tags()["key"])
		assert.Equal(t, "value2", ctx.Tags()["key2"])
		return 0
	}))
	err := L.DoString(`
		group("test", function ()
			group("nested", { key2 = "value2" }, run)
		end)
	`)
	assert.Nil(t, err)

	ctx.Logger().Info("out")
	assert.Equal(t, "value", ctx.Tags()["key"])
	assert.Equal(t, "out(key=value)", ctx.Scope())
	ctx.Exit()

	assert.Equal(t, 0, len(ctx.Tags()))
	ctx.Logger().Info("end")
	assert.Equal(t, "", ctx.Scope())

	assert.Equal(t, 5, observedLogs.Len())
	logs := observedLogs.TakeAll()
	assertLogEntry(t, logs[0], "init", nil)
	assertLogEntry(t, logs[1], "out", map[string]string{"scope": "out", "key": "value"})
	assertLogEntry(t, logs[2], "inside", map[string]string{"scope": "out.test.nested", "key": "value", "key2": "value2"})
	assertLogEntry(t, logs[3], "out", map[string]string{"scope": "out", "key": "value"})
	assertLogEntry(t, logs[4], "end", nil)
}

func assertLogEntry(t *testing.T, entry observer.LoggedEntry, message string, tags map[string]string) {
	assert.Equal(t, message, entry.Message)
	assert.Equal(t, len(tags), len(entry.Context))
	for _, f := range entry.Context {
		assert.Equal(t, zapcore.StringType, f.Type)
		assert.Equal(t, tags[f.Key], f.String)
	}
}
