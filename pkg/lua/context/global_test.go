package context_test

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	"github.com/joesonw/lte/pkg/metrics"
)

func TestGlobal(t *testing.T) {
	data := &metrics.MemoryData{}
	reporter := metrics.Memory(data)
	global := luacontext.NewGlobal(reporter)

	wg := sync.WaitGroup{}
	wg.Add(2)
	var retValue interface{}
	count := 0

	go func() {
		retValue = global.Unique("test", func() interface{} {
			count++
			return count
		})
		wg.Done()
	}()

	go func() {
		retValue = global.Unique("test", func() interface{} {
			count++
			return count
		})
		wg.Done()
	}()
	wg.Wait()

	m := metrics.Counter("test", map[string]string{"key": "value"})
	global.RegisterMetric(m)
	m.Add(2)

	assert.Equal(t, 1, count)
	assert.Equal(t, 1, retValue)

	assert.Nil(t, reporter.Finish())
	assert.Equal(t, "value", data.Counters["test"].Tags["key"])
	assert.Equal(t, int64(2), data.Counters["test"].Count)
}
