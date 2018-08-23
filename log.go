package golua

import (
	"golog"
)

var (
	_go   = "golua-go"
	goout = golog.New(_go)

	_lua   = "golua-lua"
	luaout = golog.New(_lua)
)

func init() {
	golog.SetConsoleLogByString(_go, true)
	golog.SetFileLogByString(_go, "./logs/")
	golog.SetStackByString(_go, true)
	golog.SetCallDepthByString(_go, 3)

	golog.SetConsoleLogByString(_lua, true)
	golog.SetFileLogByString(_lua, "./logs/")
}
