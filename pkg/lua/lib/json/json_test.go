package json_test

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libjson "github.com/joesonw/lte/pkg/lua/lib/json"
	test_util "github.com/joesonw/lte/pkg/lua/test-util"
)

type stringer string

func (s stringer) String() string {
	return "fmt.Stringer " + string(s)
}

type marshaler string

func (m marshaler) MarshalJSON() ([]byte, error) {
	return []byte(strconv.Quote("json.Marshaler " + string(m))), nil
}

func Test(t *testing.T) {
	test_util.Run(t, func(t *testing.T) *test_util.Test {
		return test_util.New("json", `
			a = json_marshal({
				b = true,
				n = 123,
				s = "hello",
				ud1 = ud1,
				ud2 = ud2,
				ud3 = ud3,
				table1 = { "1", "2" },
				table2 = {
					a = 1,
					b = 2,	
				},
			})
			assert(m1.b == true, "m1.b")
			assert(m1.n == 123, "m1.n")
			assert(m1.s == "hello", "m1.s")
			assert(m1.table1[1] == "1", "m1.table1[1]")
			assert(m1.table1[2] == "2", "m1.table1[2]")
			assert(m1.table2.a == 1, "m1.table2.a")
			assert(m1.table2.b == 2, "m1.table2.b")
		`).Before(func(t *testing.T, L *lua.LState, luaCtx *luacontext.Context) {
			ud1 := L.NewUserData()
			ud1.Value = stringer("ud1")
			ud2 := L.NewUserData()
			ud2.Value = marshaler("ud2")
			ud3 := L.NewUserData()
			ud3.Value = struct{}{}
			L.SetGlobal("ud1", ud1)
			L.SetGlobal("ud2", ud2)
			L.SetGlobal("ud3", ud3)
			m := map[string]interface{}{
				"b":      true,
				"n":      123,
				"s":      "hello",
				"table1": []interface{}{"1", "2"},
				"table2": map[string]interface{}{
					"a": 1,
					"b": 2,
				},
			}
			b, _ := json.Marshal(m)
			m1, err := libjson.Unmarshal(L, b)
			assert.Nil(t, err)
			L.SetGlobal("m1", m1)
		}).After(func(t *testing.T, L *lua.LState) {
			m := map[string]interface{}{
				"b":      true,
				"n":      123,
				"s":      "hello",
				"ud1":    "fmt.Stringer ud1",
				"ud2":    "json.Marshaler ud2",
				"ud3":    "*USERDATA*",
				"table1": []interface{}{"1", "2"},
				"table2": map[string]interface{}{
					"a": 1,
					"b": 2,
				},
			}
			a := L.GetGlobal("a")
			a2, _ := json.Marshal(m)
			assert.Equal(t, string(a2), a.String())
		})
	})
}
