package http_test

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	libbytes "github.com/joesonw/distress/pkg/lua/lib/bytes"
	libhttp "github.com/joesonw/distress/pkg/lua/lib/http"
	test_util "github.com/joesonw/distress/pkg/lua/test-util"
)

var testTable = []struct {
	name            string
	method          string
	url             string
	body            []byte
	headers         map[string]string
	responseStatus  int
	responseHeaders map[string]string
	responseBody    []byte
}{{
	name:            "simple http get",
	method:          "get",
	url:             "http://example.com",
	headers:         map[string]string{"Test": "123"},
	responseStatus:  200,
	responseHeaders: map[string]string{"Abc": "def"},
	responseBody:    []byte("hello world"),
}, {
	name:            "simple http get",
	method:          "post",
	url:             "http://example.com",
	body:            []byte("some content"),
	headers:         map[string]string{"Test": "123"},
	responseStatus:  200,
	responseHeaders: map[string]string{"Abc": "def"},
	responseBody:    []byte("hello world"),
}}

func Test(t *testing.T) {
	tests := make([]test_util.Testable, len(testTable))
	for i, test := range testTable {
		tests[i] = func(t *testing.T) *test_util.Test {
			client := NewTestClient(func(req *http.Request) *http.Response {
				if !strings.EqualFold(req.Method, test.method) {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("method not match, got " + req.Method + ", expected " + test.method)),
					}
				}

				if req.URL.String() != test.url {
					return &http.Response{
						StatusCode: http.StatusInternalServerError,
						Body:       ioutil.NopCloser(strings.NewReader("url not match")),
					}
				}

				if test.body != nil {
					b, err := ioutil.ReadAll(req.Body)
					if err != nil {
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       ioutil.NopCloser(strings.NewReader(err.Error())),
						}
					}
					if !bytes.Equal(b, test.body) {
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       ioutil.NopCloser(strings.NewReader("body not match")),
						}
					}
				}

				for k, v := range test.headers {
					if v != req.Header.Get(k) {
						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       ioutil.NopCloser(strings.NewReader("header " + k + " not match")),
						}
					}
				}
				res := &http.Response{Header: http.Header{}}

				for k, v := range test.responseHeaders {
					res.Header.Set(k, v)
				}
				res.StatusCode = test.responseStatus
				res.Body = ioutil.NopCloser(bytes.NewReader(test.responseBody))
				return res
			})
			var headers []string
			for k, v := range test.headers {
				headers = append(headers, fmt.Sprintf(`%s = "%s"`, k, v))
			}
			return test_util.New(test.name, fmt.Sprintf(`
					local http = require "http"
					err, body, headers, status = http:%s("%s", {
						body = "%s",
						headers = {
							%s
						},
					})()
				`, test.method, test.url, string(test.body), strings.Join(headers, ",\n"))).
				Before(func(t *testing.T, L *lua.LState, luaCtx *luacontext.Context) {
					libhttp.Open(L, luaCtx, client)
				}).
				After(func(t *testing.T, L *lua.LState) {
					assert.Equal(t, lua.LNil, L.GetGlobal("err"))
					body := libbytes.CheckValue(L, L.GetGlobal("body"))
					assert.Equal(t, string(body), string(test.responseBody))
					headers := L.GetGlobal("headers").(*lua.LTable)
					for k, v := range test.responseHeaders {
						assert.Equal(t, v, headers.RawGetString(k).String())
					}
					assert.Equal(t, test.responseStatus, int(L.GetGlobal("status").(lua.LNumber)))
				})
		}
	}
	test_util.Run(t, tests...)
}

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}
