package atom

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)
var levelMap = map[string]zapcore.Level{
	"debug": zapcore.DebugLevel,
	"info":  zapcore.InfoLevel,
	"warn":  zapcore.WarnLevel,
	"error": zapcore.ErrorLevel,
}

var Log *zap.SugaredLogger
var zaplogger *zap.Logger

func GetZapLogger() *zap.Logger {
	return zaplogger
}