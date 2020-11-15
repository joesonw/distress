package goobject

import (
	"errors"
	"fmt"
	"strings"

	lua "github.com/yuin/gopher-lua"

	libjson "github.com/joesonw/distress/pkg/lua/lib/json"
	luautil "github.com/joesonw/distress/pkg/lua/util"
)

type Object struct {
	methods    map[string]*lua.LFunction
	properties map[string]lua.LValue
	value      interface{}
}

func (obj *Object) MarshalJSON() (data []byte, err error) {
	var pairs []string

	if len(obj.properties) > 0 {
		for k, v := range obj.properties {
			data, err = libjson.Marshal(v)
			if err != nil {
				return
			}
			pairs = append(pairs, fmt.Sprintf(`"%s":%s`, k, string(data)))
		}
	}

	return []byte("{" + strings.Join(pairs, ",") + "}"), nil
}

func New(L *lua.LState, funcs map[string]lua.LGFunction, properties map[string]lua.LValue, value interface{}) lua.LValue {
	ud := L.NewUserData()
	ud.Value = value
	methods := map[string]*lua.LFunction{}
	for name, fn := range funcs {
		methods[name] = L.NewClosure(fn, ud)
	}
	return newGoObject(L, methods, properties, value)
}

func newGoObject(L *lua.LState, methods map[string]*lua.LFunction, properties map[string]lua.LValue, value interface{}) lua.LValue {
	ud := L.NewUserData()

	object := &Object{}
	object.value = value
	object.methods = methods
	object.properties = properties

	ud.Value = object

	index := L.NewTable()
	for k, v := range properties {
		index.RawSetString(k, v)
	}
	for k, f := range methods {
		index.RawSetString(k, f)
	}

	meta := L.NewTable()
	meta.RawSetString("__index", index)
	L.SetMetatable(ud, meta)
	return ud
}

func Value(value lua.LValue) (interface{}, error) {
	if value.Type() != lua.LTUserData {
		return nil, errors.New("expected user data")
	}

	ud := value.(*lua.LUserData)
	object, ok := ud.Value.(*Object)
	if !ok {
		return nil, errors.New("expected GoObject")
	}

	return object.value, nil
}

func Clone(L *lua.LState, value lua.LValue) lua.LValue {
	if value.Type() != lua.LTUserData {
		L.RaiseError("expected user data")
	}

	valueUD := value.(*lua.LUserData)
	object, ok := valueUD.Value.(*Object)
	if !ok {
		L.RaiseError("expected GoObject")
	}

	ud := L.NewUserData()
	ud.Value = value
	methods := map[string]*lua.LFunction{}
	for name, fn := range object.methods {
		methods[name] = L.NewClosure(fn.GFunction, ud)
	}

	properties := map[string]lua.LValue{}
	for key, value := range object.properties {
		properties[key] = luautil.CloneLuaValue(L, value)
	}

	return newGoObject(L, methods, properties, value)
}
