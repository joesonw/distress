package http

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/lte/pkg/lua/context"
	libasync "github.com/joesonw/lte/pkg/lua/lib/async"
	libbytes "github.com/joesonw/lte/pkg/lua/lib/bytes"
	luautil "github.com/joesonw/lte/pkg/lua/util"
	"github.com/joesonw/lte/pkg/stat"
)

const moduleName = "http"

var supportedMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodOptions,
	http.MethodTrace,
}

type httpContext struct {
	method string
	client *http.Client
	luaCtx *luacontext.Context
}

func Open(L *lua.LState, luaCtx *luacontext.Context, client *http.Client) {
	mod := L.RegisterModule(moduleName, map[string]lua.LGFunction{}).(*lua.LTable)

	for _, method := range supportedMethods {
		ud := L.NewUserData()
		ud.Value = &httpContext{
			method: method,
			client: client,
			luaCtx: luaCtx,
		}
		mod.RawSetString(strings.ToLower(method), L.NewClosure(lDo, ud))
	}
}

func lDo(L *lua.LState) int {
	c := L.CheckUserData(lua.UpvalueIndex(1)).Value.(*httpContext)
	url := L.CheckString(2)

	options := L.Get(3)
	optionsTable, _ := options.(*lua.LTable)

	var body io.Reader
	if optionsTable != nil && optionsTable.RawGetString("body") != nil {
		body = bytes.NewReader(libbytes.CheckValue(L, optionsTable.RawGetString("body")))
	}

	return libasync.DeferredResult(L, c.luaCtx.AsyncPool(), func(ctx context.Context) (lua.LGFunction, error) {
		req, err := http.NewRequest(c.method, url, body)
		if err != nil {
			return nil, err
		}

		if optionsTable != nil && optionsTable.RawGetString("headers") != nil {
			header, ok := optionsTable.RawGetString("headers").(*lua.LTable)
			if ok {
				header.ForEach(func(k, v lua.LValue) {
					req.Header.Set(k.String(), v.String())
				})
			}
		}

		start := time.Now()
		s := stat.New("http").Tag("url", url)
		defer luautil.ReportContextStat(c.luaCtx, s)
		res, err := c.client.Do(req)
		if err != nil {
			s.IntField("success", 0)
			return nil, err
		}
		defer res.Body.Close()
		s.Tag("status", strconv.Itoa(res.StatusCode))

		var b []byte
		returnResult := func(L *lua.LState) int {
			L.Push(libbytes.New(L, b))
			headers := L.NewTable()
			for k := range res.Header {
				headers.RawSetString(k, lua.LString(res.Header.Get(k)))
			}
			L.Push(headers)
			L.Push(lua.LNumber(res.StatusCode))
			return 3
		}

		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			s.IntField("success", 0)
			return returnResult, err
		}
		s.IntField("success", 1).Int64Field("duration_ns", time.Since(start).Nanoseconds())

		return returnResult, nil
	})
}
