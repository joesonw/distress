package time

import (
	"time"

	lua "github.com/yuin/gopher-lua"

	luacontext "github.com/joesonw/distress/pkg/lua/context"
	goclass "github.com/joesonw/distress/pkg/lua/lib/go-class"
)

const moduleName = "time"
const timeMetaName = "*TIME*"

type modContext struct {
	timeClass *goclass.Class
	luaCtx    *luacontext.Context
}

var modFuncs = map[string]lua.LGFunction{
	"parse": lParse,
	"now":   lNow,
}

func Open(L *lua.LState, luaCtx *luacontext.Context) {
	mod := L.RegisterModule(moduleName, map[string]lua.LGFunction{}).(*lua.LTable)
	ud := L.NewUserData()
	ud.Value = &modContext{
		timeClass: goclass.New(L, timeMetaName, timeFuncs).WithToString(L.NewFunction(timeFuncs["string"])),
		luaCtx:    luaCtx,
	}
	L.SetFuncs(mod, modFuncs, ud)
}

func New(L *lua.LState, t time.Time) lua.LValue {
	meta := L.GetTypeMetatable(timeMetaName)
	ud := L.NewUserData()
	ud.Value = t
	L.SetMetatable(ud, meta)
	return ud
}

var timeFuncs = map[string]lua.LGFunction{
	"string":     timeString,
	"format":     timeFormat,
	"unix":       timeUnix,
	"unix_nano":  timeUnixNano,
	"year":       timeYear,
	"year_day":   timeYearDay,
	"month":      timeMonth,
	"weekday":    timeWeekday,
	"day":        timeDay,
	"hour":       timeHour,
	"minute":     timeMinute,
	"second":     timeSecond,
	"nanosecond": timeNanosecond,
}

func lParse(L *lua.LState) int {
	ctx := L.CheckUserData(lua.UpvalueIndex(1)).Value.(*modContext)
	t, err := time.Parse(L.CheckString(2), L.CheckString(3))
	if err != nil {
		L.RaiseError(err.Error())
	}
	L.Push(ctx.timeClass.New(L, t))
	return 1
}

func lNow(L *lua.LState) int {
	ctx := L.CheckUserData(lua.UpvalueIndex(1)).Value.(*modContext)
	L.Push(ctx.timeClass.New(L, time.Now()))
	return 1
}

func timeString(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LString(t.String()))
	return 1
}

func timeFormat(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LString(t.Format(L.CheckString(2))))
	return 1
}

func timeUnix(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LNumber(t.Unix()))
	return 1
}

func timeUnixNano(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LNumber(t.UnixNano()))
	return 1
}

func timeYear(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LNumber(t.Year()))
	return 1
}

func timeYearDay(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LNumber(t.YearDay()))
	return 1
}

func timeMonth(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LNumber(t.Month()))
	return 1
}

func timeWeekday(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LNumber(t.Weekday()))
	return 1
}

func timeDay(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LNumber(t.Day()))
	return 1
}

func timeHour(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LNumber(t.Hour()))
	return 1
}

func timeMinute(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LNumber(t.Minute()))
	return 1
}

func timeSecond(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LNumber(t.Second()))
	return 1
}

func timeNanosecond(L *lua.LState) int {
	t := L.CheckUserData(1).Value.(time.Time)
	L.Push(lua.LNumber(t.Nanosecond()))
	return 1
}
