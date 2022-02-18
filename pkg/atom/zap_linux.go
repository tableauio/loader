//go:build linux
// +build linux

package atom

import (
	"log"
	"time"

	// rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	rotatelogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func InitZap(level string, dir string) error {
	zapLevel, ok := levelMap[level]
	if !ok {
		return errors.Errorf("illegal log level: %s", level)
	}
	writeSyncer := getWriteSyncer(dir)
	encoder := getEncoder()
	core := zapcore.NewCore(encoder, writeSyncer, zapLevel)

	zaplogger = zap.New(core, zap.AddCaller())
	Log = zaplogger.Sugar()

	Log.Infow("sugar log test1",
		"url", "http://example.com",
		"attempt", 3,
		"backoff", time.Second,
	)

	Log.Infof("sugar log test2: %s", "http://example.com")

	return nil
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.FunctionKey = "func"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.ConsoleSeparator = "|"
	return zapcore.NewConsoleEncoder(encoderConfig)
}

func getWriteSyncer(dir string) zapcore.WriteSyncer {
	// return zapcore.AddSync(&lumberjack.Logger{
	// 	Filename:   dir + "/log" + time.Now().Format("20060102"),,
	// 	MaxSize:    100, // megabytes
	// 	MaxBackups: 30,
	// 	MaxAge:     14,    //days
	// 	Compress:   false, // disabled by default
	// })

	hook, err := rotatelogs.New(
		dir+"/log"+".%Y%m%d",
		rotatelogs.WithLinkName(dir+"/log"),
		rotatelogs.WithMaxAge(time.Hour*24*7),
		rotatelogs.WithRotationTime(time.Hour*24),
	)

	if err != nil {
		log.Printf("failed to create rotatelogs: %s", err)
	}
	return zapcore.AddSync(hook)
}
