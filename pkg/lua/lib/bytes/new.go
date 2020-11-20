package bytes

import (
	"encoding/base32"
	"encoding/base64"
	"encoding/hex"

	lua "github.com/yuin/gopher-lua"
)

func lNew(L *lua.LState) int {
	s := L.CheckString(2)
	encoding := ""
	if e := L.Get(3); e != lua.LNil {
		encoding = e.String()
	}

	switch encoding {
	case "base64":
		if b, err := base64.StdEncoding.DecodeString(s); err != nil {
			L.RaiseError(err.Error())
		} else {
			L.Push(New(L, b))
		}
	case "base32":
		if b, err := base32.StdEncoding.DecodeString(s); err != nil {
			L.RaiseError(err.Error())
		} else {
			L.Push(New(L, b))
		}
	case "hex":
		if b, err := hex.DecodeString(s); err != nil {
			L.RaiseError(err.Error())
		} else {
			L.Push(New(L, b))
		}
	default:
		L.Push(New(L, []byte(s)))
	}

	return 1
}
