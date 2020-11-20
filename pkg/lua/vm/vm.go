package vm

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/spf13/afero"
	lua "github.com/yuin/gopher-lua"
	luaparse "github.com/yuin/gopher-lua/parse"
	"go.uber.org/zap"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libbase "github.com/joesonw/distress/pkg/lua/lib/base"
	libbuffer "github.com/joesonw/distress/pkg/lua/lib/buffer"
	libbytes "github.com/joesonw/distress/pkg/lua/lib/bytes"
	libcrypto "github.com/joesonw/distress/pkg/lua/lib/crypto"
	libfs "github.com/joesonw/distress/pkg/lua/lib/fs"
	libhttp "github.com/joesonw/distress/pkg/lua/lib/http"
	libjson "github.com/joesonw/distress/pkg/lua/lib/json"
	libmetrics "github.com/joesonw/distress/pkg/lua/lib/metrics"
	libnet "github.com/joesonw/distress/pkg/lua/lib/net"
	libpool "github.com/joesonw/distress/pkg/lua/lib/pool"
	libproto "github.com/joesonw/distress/pkg/lua/lib/proto"
	libtime "github.com/joesonw/distress/pkg/lua/lib/time"
	libuuid "github.com/joesonw/distress/pkg/lua/lib/uuid"
	libwebsocket "github.com/joesonw/distress/pkg/lua/lib/websocket"
)

var idCounter int64

func Compile(src, name string) (*lua.FunctionProto, error) {
	chunk, err := luaparse.Parse(strings.NewReader(src), name)
	if err != nil {
		return nil, err
	}

	return lua.Compile(chunk, name)
}

type VM struct {
	*sync.Mutex
	id          int64
	params      Parameters
	state       *lua.LState
	logger      *zap.Logger
	releasePool *libpool.ReleasePool
	asyncPool   *libpool.AsyncPool
	fn          *lua.LFunction
}

type Parameters struct {
	AsyncPoolConcurrency int
	AsyncPoolTimeout     time.Duration
	AsyncPoolBufferSize  int
	EnvVars              map[string]string
	Filesystem           afero.Fs
}

func New(logger *zap.Logger, global *luacontext.Global, params Parameters) *VM {
	id := atomic.AddInt64(&idCounter, 1)
	logger = logger.With(zap.Int64("vm_id", id))

	if params.AsyncPoolConcurrency <= 0 {
		params.AsyncPoolConcurrency = 4
	}

	if params.AsyncPoolBufferSize <= 0 || params.AsyncPoolBufferSize < params.AsyncPoolConcurrency {
		params.AsyncPoolBufferSize = params.AsyncPoolConcurrency * 16
	}

	asyncPool := libpool.NewAsync(logger, params.AsyncPoolConcurrency, params.AsyncPoolTimeout, params.AsyncPoolBufferSize)
	releasePool := libpool.NewRelease(logger)
	asyncPool.Start()

	L := lua.NewState(lua.Options{
		CallStackSize:       0,
		RegistrySize:        0,
		RegistryMaxSize:     0,
		RegistryGrowStep:    0,
		SkipOpenLibs:        true,
		IncludeGoStackTrace: false,
		MinimizeStackMemory: false,
	})

	luaCtx := luacontext.New(L, global, releasePool, asyncPool, logger)

	lua.OpenBase(L)
	lua.OpenPackage(L)
	lua.OpenMath(L)
	lua.OpenString(L)
	lua.OpenTable(L)
	lua.OpenOs(L)

	libjson.Open(L, luaCtx)
	libbytes.Open(L, luaCtx)
	libtime.Open(L, luaCtx)
	libbase.Open(L, luaCtx, params.Filesystem)

	libfs.Open(L, luaCtx, params.Filesystem)
	libhttp.Open(L, luaCtx, &http.Client{})
	libproto.Open(L, luaCtx, params.Filesystem)
	libwebsocket.Open(L, luaCtx)
	libnet.Open(L, luaCtx)
	libuuid.Open(L, luaCtx)
	libcrypto.Open(L, luaCtx)
	libmetrics.Open(L, luaCtx)
	libbuffer.Open(L, luaCtx)

	for k, v := range params.EnvVars {
		L.Env.RawSetString(k, lua.LString(v))
	}

	vm := &VM{
		Mutex:       &sync.Mutex{},
		id:          id,
		params:      params,
		logger:      logger,
		state:       L,
		asyncPool:   asyncPool,
		releasePool: releasePool,
	}
	return vm
}

func (vm *VM) Load(proto *lua.FunctionProto) error {
	vm.state.Push(vm.state.NewFunctionFromProto(proto))
	if err := vm.state.PCall(0, lua.MultRet, nil); err != nil {
		return err
	}

	val := vm.state.GetGlobal("run")
	fn, ok := val.(*lua.LFunction)
	if !ok {
		return fmt.Errorf("expect global function run()")
	}

	vm.fn = fn
	return nil
}

func (vm *VM) ID() int64 {
	return vm.id
}

func (vm *VM) Run(id int64) error {
	vm.state.Push(vm.fn)
	vm.state.Push(lua.LNumber(id))
	return vm.state.PCall(1, 0, nil)
}

func (vm *VM) LState() *lua.LState {
	return vm.state
}

func (vm *VM) Reset() {
	vm.releasePool.Clean()
	vm.asyncPool.Stop()
	vm.asyncPool.Start()
}
