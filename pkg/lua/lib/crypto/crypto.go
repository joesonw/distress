package crypto

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	lua "github.com/yuin/gopher-lua"
	"golang.org/x/crypto/sha3"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libbytes "github.com/joesonw/lte/pkg/lua/lib/bytes"
)

const moduleName = "crypto"

func Open(L *lua.LState, luaCtx *luacontext.Context) {
	mod := L.RegisterModule(moduleName, map[string]lua.LGFunction{}).(*lua.LTable)

	mod.RawSetString("md5", newHash(L, md5.New()))
	mod.RawSetString("sha1", newHash(L, sha1.New()))
	mod.RawSetString("sha256", newHash(L, sha256.New()))
	mod.RawSetString("sha512", newHash(L, sha512.New()))
	mod.RawSetString("sha3_224", newHash(L, sha3.New224()))
	mod.RawSetString("sha3_256", newHash(L, sha3.New256()))
	mod.RawSetString("sha3_384", newHash(L, sha3.New384()))
	mod.RawSetString("sha3_512", newHash(L, sha3.New512()))
}

func newHash(L *lua.LState, h hash.Hash) *lua.LFunction {
	ud := L.NewUserData()
	ud.Value = h
	return L.NewClosure(lHash, ud)
}

func lHash(L *lua.LState) int {
	h := L.CheckUserData(lua.UpvalueIndex(1)).Value.(hash.Hash)
	source := libbytes.Check(L, 2)
	L.Push(libbytes.New(L, h.Sum(source)))
	return 1
}
