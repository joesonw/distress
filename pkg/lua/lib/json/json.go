package json

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	luautil "github.com/joesonw/lte/pkg/lua/util"
)

var (
	errNested      = errors.New("cannot encode recursively nested tables to JSON")
	errSparseArray = errors.New("cannot encode sparse array")
	errInvalidKeys = errors.New("cannot encode mixed or invalid key types")
	funcs          = map[string]lua.LGFunction{
		"json_unmarshal": lUnmarshal,
		"json_marshal":   lMarshal,
	}
)

type errInvalidType lua.LValueType

func (e errInvalidType) Error() string {
	return `cannot encode ` + lua.LValueType(e).String() + ` to JSON`
}

func IsErrInvalidType(err error) bool {
	return errors.Is(err, errInvalidType(lua.LTNil))
}

func Marshal(value lua.LValue) ([]byte, error) {
	return json.Marshal(jsonValue{
		LValue:  value,
		visited: make(map[*lua.LTable]bool),
	})
}

type jsonValue struct {
	lua.LValue
	visited map[*lua.LTable]bool
}

func (j jsonValue) MarshalJSON() (data []byte, err error) {
	switch converted := j.LValue.(type) {
	case lua.LBool:
		data, err = json.Marshal(bool(converted))
	case lua.LNumber:
		data, err = json.Marshal(float64(converted))
	case *lua.LNilType:
		data = []byte(`null`)
	case lua.LString:
		data, err = json.Marshal(string(converted))
	case *lua.LUserData:
		if stringer, ok := converted.Value.(fmt.Stringer); ok {
			data = []byte(`"` + stringer.String() + `"`)
		} else if marshaller, ok := converted.Value.(json.Marshaler); ok {
			data, err = marshaller.MarshalJSON()
		} else {
			data = []byte(`"*USERDATA*"`)
		}
	case *lua.LTable:
		if j.visited[converted] {
			return nil, errNested
		}
		j.visited[converted] = true

		key, value := converted.Next(lua.LNil)

		switch key.Type() {
		case lua.LTNil: // empty table
			data = []byte(`[]`)
		case lua.LTNumber:
			arr := make([]jsonValue, 0, converted.Len())
			expectedKey := lua.LNumber(1)
			for key != lua.LNil {
				if key.Type() != lua.LTNumber {
					err = errInvalidKeys
					return
				}
				if expectedKey != key {
					err = errSparseArray
					return
				}
				arr = append(arr, jsonValue{value, j.visited})
				expectedKey++
				key, value = converted.Next(key)
			}
			data, err = json.Marshal(arr)
		case lua.LTString:
			obj := make(map[string]jsonValue)
			for key != lua.LNil {
				if key.Type() != lua.LTString {
					err = errInvalidKeys
					return
				}
				obj[key.String()] = jsonValue{value, j.visited}
				key, value = converted.Next(key)
			}
			data, err = json.Marshal(obj)
		default:
			err = errInvalidKeys
		}
	default:
		err = errInvalidType(j.LValue.Type())
	}
	return
}

func Unmarshal(L *lua.LState, data []byte) (lua.LValue, error) {
	var value interface{}
	err := json.Unmarshal(data, &value)
	if err != nil {
		return nil, err
	}
	return UnmarshalGoValue(L, value), nil
}

func UnmarshalGoValue(L *lua.LState, value interface{}) lua.LValue {
	if value == nil {
		return lua.LNil
	}
	switch converted := value.(type) {
	case bool:
		return lua.LBool(converted)
	case float64:
		return lua.LNumber(converted)
	case int64:
		return lua.LNumber(converted)
	case time.Time:
		return lua.LNumber(converted.UnixNano() / 1000000)
	case string:
		return lua.LString(converted)
	case json.Number:
		return lua.LString(converted)
	case []interface{}:
		arr := L.CreateTable(len(converted), 0)
		for _, item := range converted {
			arr.Append(UnmarshalGoValue(L, item))
		}
		return arr
	case map[string]interface{}:
		tbl := L.CreateTable(0, len(converted))
		for key, item := range converted {
			tbl.RawSetH(lua.LString(key), UnmarshalGoValue(L, item))
		}
		return tbl
	case nil:
		return lua.LNil
	}

	return lua.LNil
}

func Open(L *lua.LState, luaCtx *luacontext.Context) {
	luautil.RegisterGlobalFuncs(L, funcs)
}

func lUnmarshal(L *lua.LState) int {
	if L.GetTop() != 1 {
		L.ArgError(1, "json_unmarshal(string): takes one argument")
		return 0
	}

	val, err := Unmarshal(L, []byte(L.Get(1).String()))
	if err != nil {
		L.ArgError(1, err.Error())
		return 0
	}

	L.Push(val)
	return 1
}

func lMarshal(L *lua.LState) int {
	if L.GetTop() != 1 {
		L.ArgError(1, "json_marshal(table): takes one argument")
		return 0
	}

	bytes, err := Marshal(L.Get(1))
	if err != nil {
		L.ArgError(1, err.Error())
		return 0
	}

	L.Push(lua.LString(bytes))
	return 1
}
