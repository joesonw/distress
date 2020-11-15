package value

import (
	"errors"
	"reflect"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
	"go.uber.org/multierr"
)

type Marshaler interface {
	MarshalLua(L *lua.LState) (lua.LValue, error)
}

type Unmarshaler interface {
	UnmarshalLua(value lua.LValue, in interface{}) error
}

var _typeTimeDuration = reflect.TypeOf(time.Duration(0))
var _typeLuaFunction = reflect.TypeOf((*lua.FunctionProto)(nil))
var _typeMarshaler = reflect.TypeOf((*Marshaler)(nil)).Elem()
var _typeUnmarshaler = reflect.TypeOf((*Unmarshaler)(nil)).Elem()

func Unmarshal(lValue lua.LValue, in interface{}) error {
	val := reflect.ValueOf(in)
	return unmarshalValue(val.Type(), val, lValue, nil)
}

func unmarshalTable(table *lua.LTable, val reflect.Value) error {
	if val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	n := val.NumField()
	typ := val.Type()

	for i := 0; i < n; i++ {
		fieldType := typ.Field(i)
		field := val.Field(i)
		jsonTag := fieldType.Tag.Get("json")
		tags := strings.Split(jsonTag, ",")
		jsonName := tags[0]

		if jsonName == "-" {
			continue
		}

		lVal := table.RawGetString(jsonName)
		if err := unmarshalValue(fieldType.Type, field, lVal, tags); err != nil {
			return err
		}
	}

	return nil
}

func unmarshalValue(typ reflect.Type, field reflect.Value, value lua.LValue, tags []string) error {
	if field.Type().Implements(_typeUnmarshaler) {
		return field.Interface().(Unmarshaler).UnmarshalLua(value, field.Interface())
	}

	if value.Type() == lua.LTNil {
		return nil
	}

	if typ == _typeTimeDuration {
		if value.Type() != lua.LTString {
			return errors.New(value.Type().String() + " cannot be assigned to " + _typeTimeDuration.String())
		}

		dur, err := time.ParseDuration(value.String())
		if err != nil {
			return err
		}

		field.Set(reflect.ValueOf(dur))
		return nil
	}

	if typ == _typeLuaFunction {
		if value.Type() != lua.LTFunction {
			return errors.New(value.Type().String() + " cannot be assigned to " + _typeLuaFunction.String())
		}

		fn := value.(*lua.LFunction)
		field.Set(reflect.ValueOf(fn.Proto))
		return nil
	}

	isUserData := false
	for _, tag := range tags {
		if tag == "UserData" {
			isUserData = true
			break
		}
	}

	kind := typ.Kind()
	switch kind {
	case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint, reflect.Float64, reflect.Float32:
		if value.Type() != lua.LTNumber {
			return errors.New(value.Type().String() + " cannot be assigned to " + kind.String())
		}
	case reflect.String:
		if value.Type() != lua.LTString {
			return errors.New(value.Type().String() + " cannot be assigned to " + kind.String())
		}
	case reflect.Bool:
		if value.Type() != lua.LTBool {
			return errors.New(value.Type().String() + " cannot be assigned to " + kind.String())
		}
	case reflect.Map, reflect.Slice:
		if value.Type() != lua.LTTable {
			return errors.New(value.Type().String() + " cannot be assigned to " + kind.String())
		}
	case reflect.Ptr:
		if isUserData && value.Type() != lua.LTUserData {
			return errors.New(value.Type().String() + " cannot be assigned to userdata")
		} else if !isUserData && value.Type() != lua.LTTable {
			return errors.New(value.Type().String() + " cannot be assigned to " + kind.String())
		}
	default:
		return errors.New(value.Type().String() + " cannot be assigned to " + kind.String())
	}

	switch kind {
	case reflect.Int64:
		field.Set(reflect.ValueOf(int64(value.(lua.LNumber))))
	case reflect.Int32:
		field.Set(reflect.ValueOf(int32(value.(lua.LNumber))))
	case reflect.Int16:
		field.Set(reflect.ValueOf(int16(value.(lua.LNumber))))
	case reflect.Int:
		field.Set(reflect.ValueOf(int(value.(lua.LNumber))))
	case reflect.Uint64:
		field.Set(reflect.ValueOf(uint64(value.(lua.LNumber))))
	case reflect.Uint32:
		field.Set(reflect.ValueOf(uint32(value.(lua.LNumber))))
	case reflect.Uint16:
		field.Set(reflect.ValueOf(uint16(value.(lua.LNumber))))
	case reflect.Uint:
		field.Set(reflect.ValueOf(uint(value.(lua.LNumber))))
	case reflect.Float64:
		field.Set(reflect.ValueOf(float64(value.(lua.LNumber))))
	case reflect.Float32:
		field.Set(reflect.ValueOf(float32(value.(lua.LNumber))))
	case reflect.String:
		field.Set(reflect.ValueOf(value.String()))
	case reflect.Bool:
		field.Set(reflect.ValueOf(bool(value.(lua.LBool))))
	case reflect.Map:
		{
			table := value.(*lua.LTable)
			typ := field.Type()
			keyType := typ.Key()
			valueType := typ.Elem()
			var m reflect.Value
			if field.CanSet() {
				m = reflect.MakeMap(field.Type())
			} else {
				m = field
			}
			var errs []error
			table.ForEach(func(k, v lua.LValue) {
				key := reflect.New(keyType).Elem()
				value := reflect.New(valueType).Elem()
				errs = append(errs, unmarshalValue(keyType, key, k, tags))
				errs = append(errs, unmarshalValue(valueType, value, v, tags))
				m.SetMapIndex(key, value)
			})
			if err := multierr.Combine(errs...); err != nil {
				return err
			}
			if field.CanSet() {
				field.Set(m)
			}
		}
	case reflect.Slice:
		{
			table := value.(*lua.LTable)
			n := table.Len()
			childType := field.Type().Elem()
			slice := reflect.MakeSlice(field.Type(), n, n)
			for i := 0; i < n; i++ {
				child := reflect.New(childType).Elem()
				if err := unmarshalValue(childType, child, table.RawGetInt(i+1), tags); err != nil {
					return err
				}
				slice.Index(i).Set(child)
			}
			field.Set(slice)
		}
	case reflect.Ptr:
		if isUserData {
			ud := value.(*lua.LUserData)
			field.Set(reflect.ValueOf(ud.Value))
		} else {
			return unmarshalTable(value.(*lua.LTable), field)
		}
	}

	return nil
}

func Marshal(L *lua.LState, in interface{}) (lua.LValue, error) {
	if in == nil {
		return lua.LNil, nil
	}
	val := reflect.ValueOf(in)
	return marshalValue(L, val.Type(), val, nil)
}

func marshalToTable(L *lua.LState, val reflect.Value) (*lua.LTable, error) {
	table := L.NewTable()
	if val.Type().Kind() == reflect.Ptr {
		val = val.Elem()
	}
	n := val.NumField()
	typ := val.Type()

	for i := 0; i < n; i++ {
		fieldType := typ.Field(i)
		field := val.Field(i)
		jsonTag := fieldType.Tag.Get("json")
		tags := strings.Split(jsonTag, ",")
		jsonName := tags[0]
		if jsonName == "-" {
			continue
		}

		lVal, err := marshalValue(L, fieldType.Type, field, tags)
		if err != nil {
			return nil, err
		}
		table.RawSetString(jsonName, lVal)
	}

	return table, nil
}

func marshalValue(L *lua.LState, typ reflect.Type, field reflect.Value, tags []string) (lua.LValue, error) {
	if typ.Implements(_typeMarshaler) {
		return field.Interface().(Marshaler).MarshalLua(L)
	}

	if typ == _typeTimeDuration {
		dur := field.Interface().(time.Duration)
		return lua.LString(dur.String()), nil
	}

	if typ == _typeLuaFunction {
		proto := field.Interface().(*lua.FunctionProto)
		return L.NewFunctionFromProto(proto), nil
	}

	isUserData := false
	for _, tag := range tags {
		if tag == "UserData" {
			isUserData = true
			break
		}
	}

	kind := typ.Kind()
	switch kind {
	case reflect.Int64:
		return lua.LNumber(field.Interface().(int64)), nil
	case reflect.Int32:
		return lua.LNumber(field.Interface().(int32)), nil
	case reflect.Int16:
		return lua.LNumber(field.Interface().(int16)), nil
	case reflect.Int:
		return lua.LNumber(field.Interface().(int)), nil
	case reflect.Uint64:
		return lua.LNumber(field.Interface().(uint64)), nil
	case reflect.Uint32:
		return lua.LNumber(field.Interface().(uint32)), nil
	case reflect.Uint16:
		return lua.LNumber(field.Interface().(uint16)), nil
	case reflect.Uint:
		return lua.LNumber(field.Interface().(uint)), nil
	case reflect.Float64:
		return lua.LNumber(field.Interface().(float64)), nil
	case reflect.Float32:
		return lua.LNumber(field.Interface().(float32)), nil
	case reflect.String:
		return lua.LString(field.Interface().(string)), nil
	case reflect.Bool:
		return lua.LBool(field.Interface().(bool)), nil
	case reflect.Map:
		{
			table := L.NewTable()
			typ := field.Type()
			keyType := typ.Key()
			valueType := typ.Elem()
			iter := field.MapRange()
			for iter.Next() {
				key, err := marshalValue(L, keyType, iter.Key(), tags)
				if err != nil {
					return nil, err
				}
				value, err := marshalValue(L, valueType, iter.Value(), tags)
				if err != nil {
					return nil, err
				}
				table.RawSet(key, value)
			}
			return table, nil
		}
	case reflect.Slice:
		{
			table := L.NewTable()
			n := field.Len()
			typ := field.Type().Elem()
			for i := 0; i < n; i++ {
				lVal, err := marshalValue(L, typ, field.Index(i), tags)
				if err != nil {
					return nil, err
				}
				table.RawSetInt(i+1, lVal)
			}
			return table, nil
		}
	case reflect.Ptr:
		if isUserData {
			ud := L.NewUserData()
			ud.Value = field.Interface()
			return ud, nil
		} else {
			return marshalToTable(L, field)
		}
	}
	return nil, errors.New(kind.String() + " cannot be casted to lua")
}

func UnmarshalTo(L *lua.LState, lValue lua.LValue) (interface{}, error) {
	switch lValue.Type() {
	case lua.LTBool:
		return lValue.(lua.LBool), nil
	case lua.LTString:
		return lValue.(lua.LString), nil
	case lua.LTNumber:
		return lValue.(lua.LNumber), nil
	case lua.LTTable:
		table := lValue.(*lua.LTable)
		if length := table.Len(); length > 0 { // is array
			arr := make([]interface{}, length)
			var err error
			for i := 0; i < length; i++ {
				arr[i], err = UnmarshalTo(L, table.RawGetInt(i+1))
				if err != nil {
					return nil, err
				}
			}
			return arr, nil
		} else { // is map
			m := map[string]interface{}{}
			var errs []error
			table.ForEach(func(k, v lua.LValue) {
				in, err := UnmarshalTo(L, v)
				if err != nil {
					errs = append(errs, err)
					return
				}

				m[k.String()] = in
			})
			err := multierr.Combine(errs...)
			if err != nil {
				return nil, err
			}
			return m, nil
		}
	}
	return nil, errors.New(lValue.Type().String() + " is not supported")
}
