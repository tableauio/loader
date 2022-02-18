//go:build windows
// +build windows

package atom

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func InitZap(level string, dir string) error {
	logFile := "log.log"
	zapLevel, ok := levelMap[level]
	if !ok {
		log.Fatalf("illegal log level: %s", level)
		return errors.New("illegal log level")
	}
	zap.RegisterSink("lumberjack", func(*url.URL) (zap.Sink, error) {
		return lumberjackSink{
			Logger: getLumberjackLogger(dir),
		}, nil
	})
	loggerConfig := zap.Config{
		Level:         zap.NewAtomicLevelAt(zapLevel),
		Development:   true,
		Encoding:      "console",
		EncoderConfig: getEncoderConfig(),
		OutputPaths:   []string{fmt.Sprintf("lumberjack:%s", logFile)},
	}
	var err error
	zaplogger, err = loggerConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("build zap logger from config error: %v", err))
	}
	zap.ReplaceGlobals(zaplogger)
	Log = zaplogger.Sugar() // NewSugar("sugar")

	Log.Infow("sugar log test1",
		"url", "http://example.com",
		"attempt", 3,
		"backoff", time.Second,
	)

	Log.Infof("sugar log test2: %s", "http://example.com")

	return nil
}

func NewSugar(name string) *zap.SugaredLogger {
	return zaplogger.Named(name).Sugar()
}

func getEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:          "ts",
		LevelKey:         "level",
		NameKey:          "logger",
		CallerKey:        "caller",
		MessageKey:       "msg",
		StacktraceKey:    "stacktrace",
		FunctionKey:      "func",
		ConsoleSeparator: "|",
		LineEnding:       zapcore.DefaultLineEnding,
		EncodeLevel:      zapcore.CapitalLevelEncoder,
		EncodeTime:       zapcore.ISO8601TimeEncoder,
		EncodeDuration:   zapcore.SecondsDurationEncoder,
		EncodeCaller:     zapcore.ShortCallerEncoder,
	}
}

type lumberjackSink struct {
	*lumberjack.Logger
}

func (lumberjackSink) Sync() error {
	return nil
}

func getLumberjackLogger(dir string) *lumberjack.Logger {
	return &lumberjack.Logger{
		Filename:   "log.log",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     28,   //days
		Compress:   true, // disabled by default
	}
}
