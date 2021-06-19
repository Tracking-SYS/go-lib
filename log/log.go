// provide a global opinionated logger using zap
package log

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap/zapcore"
)

type Logger = logr.Logger

var Discard = logr.Discard

var (
	Log   = Factory("Smarter")
	Warn  = WARN()
	Debug = DEBUG()
)

// logr level is invert of zap log level
// see more https://github.com/go-logr/zapr#implementation-details
const (
	LogDebugLevel int = -int(zapcore.DebugLevel)
	LogInfoLevel      = -int(zapcore.InfoLevel)
	LogWarnLevel      = -int(zapcore.WarnLevel)
	LogErrorLevel     = -int(zapcore.ErrorLevel)
	LogFatalLevel     = -int(zapcore.FatalLevel)
)

func CloseHook() {
	_ = global.Sync()
}

func Factory(name string) Logger {
	return zapr.NewLogger(global).WithName(name)
}

// global instance
var (
	logDefault   = Factory("ko")
	warnDefault  = logDefault.V(LogWarnLevel)
	debugDefault = logDefault.V(LogDebugLevel)
)

func Fatal(err error, msg string, keysAndValues ...interface{}) {
	logDefault.Error(err, msg, keysAndValues...)
	panic(err)
}

// the default logger use Level INFO
func LOG() Logger {
	return logDefault
}

// the default logger use Level WARN
func WARN() Logger {
	return warnDefault
}

// the default logger use Level DEBUG
func DEBUG() Logger {
	return debugDefault
}
