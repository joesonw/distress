package context

import (
	"fmt"
	"strings"
	"sync"

	lua "github.com/yuin/gopher-lua"
	"go.uber.org/zap"

	libpool "github.com/joesonw/distress/pkg/lua/lib/pool"
)

type scope struct {
	name string
	tags map[string]string
}

type Context struct {
	global      *Global
	mu          *sync.Mutex
	logger      *zap.Logger
	releasePool *libpool.ReleasePool
	asyncPool   *libpool.AsyncPool
	scopes      []scope
}

func New(L *lua.LState, global *Global, releasePool *libpool.ReleasePool, asyncPool *libpool.AsyncPool, logger *zap.Logger) *Context {
	c := &Context{
		mu: &sync.Mutex{}, releasePool: releasePool,
		asyncPool: asyncPool,
		logger:    logger,
		global:    global,
	}

	ud := L.NewUserData()
	ud.Value = c
	L.SetGlobal("group", L.NewClosure(lGroup, ud))

	return c
}

func lGroup(L *lua.LState) int {
	c := L.CheckUserData(lua.UpvalueIndex(1)).Value.(*Context)
	var lTags *lua.LTable
	var lFn *lua.LFunction
	var lName string
	arg1 := L.Get(1)

	switch arg1.Type() {
	case lua.LTTable:
		lTags = arg1.(*lua.LTable)
		lFn = L.CheckFunction(2)
	case lua.LTString:
		lName = arg1.String()
		arg2 := L.Get(2)
		if arg2.Type() == lua.LTTable {
			lTags = arg2.(*lua.LTable)
			lFn = L.CheckFunction(3)
		} else {
			lFn = L.CheckFunction(2)
		}
	default:
		L.RaiseError("group(name, function) or group(name, tags, function) or group(tags, function)")
	}

	var tagKeyValues []string
	if lTags != nil {
		lTags.ForEach(func(k, v lua.LValue) {
			tagKeyValues = append(tagKeyValues, k.String(), v.String())
		})
	}
	c.Enter(lName, tagKeyValues...)
	err := L.CallByParam(lua.P{
		Fn:      lFn,
		Protect: true,
	})
	c.Exit()
	if err != nil {
		L.RaiseError(err.Error())
	}
	return 0
}

func (c *Context) ReleasePool() *libpool.ReleasePool {
	return c.releasePool
}

func (c *Context) AsyncPool() *libpool.AsyncPool {
	return c.asyncPool
}

func (c *Context) Logger() *zap.Logger {
	logger := c.logger
	var names []string
	for i := range c.scopes {
		if c.scopes[i].name != "" {
			names = append(names, c.scopes[i].name)
		}
		if len(c.scopes[i].tags) > 0 {
			for k, v := range c.scopes[i].tags {
				logger = logger.With(zap.String(k, v))
			}
		}
	}
	if len(names) > 0 {
		logger = logger.With(zap.String("scope", strings.Join(names, ".")))
	}
	return logger
}

func (c *Context) Tags() map[string]string {
	tags := map[string]string{}
	var names []string
	for i := range c.scopes {
		if c.scopes[i].name != "" {
			names = append(names, c.scopes[i].name)
		}
		if len(c.scopes[i].tags) > 0 {
			for k, v := range c.scopes[i].tags {
				tags[k] = v
			}
		}
	}
	if len(names) > 0 {
		tags["scope"] = strings.Join(names, ".")
	}
	return tags
}

func (c *Context) Scope() string {
	var names []string
	var tags []string
	for i := range c.scopes {
		if c.scopes[i].name != "" {
			names = append(names, c.scopes[i].name)
		}
		if len(c.scopes[i].tags) > 0 {
			for k, v := range c.scopes[i].tags {
				tags = append(tags, fmt.Sprintf("%s=%s", k, v))
			}
		}
	}
	s := ""
	if len(names) > 0 {
		s += strings.Join(names, ".")
	}
	if len(tags) > 0 {
		s += "(" + strings.Join(tags, ",") + ")"
	}
	return s
}

func (c *Context) Enter(name string, tagKeyValues ...string) *Context {
	c.mu.Lock()
	defer c.mu.Unlock()
	n := len(tagKeyValues)
	s := scope{
		name: name,
		tags: map[string]string{},
	}
	for i := 0; i < n; i += 2 {
		s.tags[tagKeyValues[i]] = tagKeyValues[i+1]
	}
	c.scopes = append(c.scopes, s)
	return c
}

func (c *Context) Exit() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.scopes = c.scopes[:len(c.scopes)-1]
}

func (c *Context) Global() *Global {
	return c.global
}
