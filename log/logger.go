package log

import (
	"os"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// TODO: implement a mechanism to load this values from configuration
const (
	logFilePath = "yhs.log"
	maxSize     = 5
	maxBackups  = 10
	maxAge      = 14
	compress    = true
	logLevel    = zap.InfoLevel
)

var (
	once   sync.Once
	Logger *zap.Logger
)

func init() {
	once.Do(func() {
		stdout := zapcore.AddSync(os.Stdout)
		file := zapcore.AddSync(&lumberjack.Logger{
			Filename:   logFilePath,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   compress,
		})

		productionCfg := zap.NewProductionEncoderConfig()
		productionCfg.TimeKey = "timestamp"
		productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		developmentCfg := zap.NewDevelopmentEncoderConfig()
		developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

		consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
		fileEncoder := zapcore.NewJSONEncoder(productionCfg)

		core := zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, stdout, logLevel),
			zapcore.NewCore(fileEncoder, file, logLevel),
		)

		Logger = zap.New(core)
	})
}
