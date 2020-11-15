package proto

import (
	protodesc "github.com/jhump/protoreflect/desc"
	"github.com/jhump/protoreflect/dynamic"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libbytes "github.com/joesonw/distress/pkg/lua/lib/bytes"
	libjson "github.com/joesonw/distress/pkg/lua/lib/json"
)

const messageMetaName = "*PROTO*MESSAGE*"

type messageContext struct {
	desc   *protodesc.MessageDescriptor
	luaCtx *luacontext.Context
}

var messageFuncs = map[string]lua.LGFunction{
	"encode": messageEncode,
	"decode": messageDecode,
}

func messageEncode(L *lua.LState) int {
	c := L.CheckUserData(1).Value.(*messageContext)
	obj := L.CheckTable(2)
	bytes, err := libjson.Marshal(obj)
	if err != nil {
		L.RaiseError(err.Error())
	}

	message := dynamic.NewMessage(c.desc)
	if err := message.UnmarshalJSON(bytes); err != nil {
		L.RaiseError(err.Error())
	}

	bytes, err = message.Marshal()
	if err != nil {
		L.RaiseError(err.Error())
	}
	L.Push(libbytes.New(L, bytes))
	return 1
}

func messageDecode(L *lua.LState) int {
	c := L.CheckUserData(1).Value.(*messageContext)
	bytes := libbytes.Check(L, 2)

	message := dynamic.NewMessage(c.desc)
	err := message.Unmarshal(bytes)
	if err != nil {
		L.RaiseError(err.Error())
	}

	bytes, err = message.MarshalJSON()
	if err != nil {
		L.RaiseError(err.Error())
	}

	val, err := libjson.Unmarshal(L, bytes)
	if err != nil {
		L.RaiseError(err.Error())
	}

	L.Push(val)
	return 1
}
