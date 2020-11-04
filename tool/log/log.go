package log

import (
	"strings"

	"github.com/astaxie/beego/logs"
)

var (
	Logger *logs.BeeLogger
)

// 初始化日志
func Init(filename string, level string) {
	//logs.SetLogger("file")
	Logger = logs.NewLogger(10000)

	Logger.EnableFuncCallDepth(true)
	Logger.SetLogger("file",   `{"filename":"test.log"}`)
	Logger.SetLogger("console", "")
	Logger.SetLogFuncCallDepth(2)

	var logLevel int
	switch strings.ToLower(level) {
	case "trace":
		logLevel = logs.LevelTrace
	case "debug":
		logLevel = logs.LevelDebug
	case "info":
		logLevel = logs.LevelInfo
	case "warn":
		logLevel = logs.LevelWarn
	case "error":
		logLevel = logs.LevelError
	case "critical":
		logLevel = logs.LevelCritical
	default:
		logLevel = logs.LevelWarn
	}

	Logger.SetLevel(logLevel)
}
