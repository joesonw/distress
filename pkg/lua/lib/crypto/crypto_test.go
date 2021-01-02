package crypto_test

import (
	"bytes"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"
	"golang.org/x/crypto/sha3"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libbytes "github.com/joesonw/lte/pkg/lua/lib/bytes"
	libcrypto "github.com/joesonw/lte/pkg/lua/lib/crypto"
	test_util "github.com/joesonw/lte/pkg/lua/test-util"
)

var testTable = []struct {
	name string
	hash hash.Hash
}{{
	name: "md5",
	hash: md5.New(),
}, {
	name: "sha1",
	hash: sha1.New(),
}, {
	name: "sha256",
	hash: sha256.New(),
}, {
	name: "sha512",
	hash: sha512.New(),
}, {
	name: "sha3_224",
	hash: sha3.New224(),
}, {
	name: "sha3_256",
	hash: sha3.New256(),
}, {
	name: "sha3_384",
	hash: sha3.New384(),
}, {
	name: "sha3_512",
	hash: sha3.New512(),
}}

func Test(t *testing.T) {
	tests := make([]test_util.Testable, len(testTable))
	for i := range testTable {
		test := testTable[i]
		tests[i] = func(_ *testing.T) *test_util.Test {
			source := make([]byte, 8)
			_, err := rand.Read(source)
			assert.Nil(t, err)

			println(fmt.Sprintf(`
					local crypto = require("crypto")
					return crypto:%s(data)
				`, test.name))
			return test_util.New(test.name, fmt.Sprintf(`
					local crypto = require("crypto")
					result = crypto:%s(data)
				`, test.name)).
				Before(func(t *testing.T, L *lua.LState, luaCtx *luacontext.Context) {
					libcrypto.Open(L, luaCtx)
					L.SetGlobal("data", libbytes.New(L, source))
				}).
				After(func(t *testing.T, L *lua.LState) {
					result := libbytes.CheckValue(L, L.GetGlobal("result"))
					expected := test.hash.Sum(source)
					assert.True(t, bytes.Equal(expected, result))
				})
		}
	}
	test_util.Run(t, tests...)
}
