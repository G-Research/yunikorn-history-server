package log

import (
	"os"
	"strconv"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LogConfig struct {
	LogFilePath string
	MaxSize     int
	MaxBackups  int
	MaxAge      int
	Compress    bool
	LogLevel    string
}

var (
	once   sync.Once
	Logger *zap.Logger
)

func InitLogger(config LogConfig) {
	once.Do(func() {
		stdout := zapcore.AddSync(os.Stdout)
		file := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.LogFilePath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		})

		productionCfg := zap.NewProductionEncoderConfig()
		productionCfg.TimeKey = "timestamp"
		productionCfg.EncodeTime = zapcore.ISO8601TimeEncoder

		developmentCfg := zap.NewDevelopmentEncoderConfig()
		developmentCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder

		consoleEncoder := zapcore.NewConsoleEncoder(developmentCfg)
		fileEncoder := zapcore.NewJSONEncoder(productionCfg)

		core := zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, stdout, parseLevel(config.LogLevel)),
			zapcore.NewCore(fileEncoder, file, parseLevel(config.LogLevel)),
		)

		Logger = zap.New(core)
	})
}

// parseLevel parses a textual (or numeric) log level into a `zapcore.Level` instance.
// Both numeric (-1 <= level <= 5)
// and textual (DEBUG, INFO, WARN, ERROR, DPANIC, PANIC, FATAL) are supported.
// Ref: https://github.com/apache/yunikorn-core/blob/a786feb5761be28e802d08976d224c40639cd86b/pkg/log/logger.go#L301
func parseLevel(level string) *zapcore.Level {
	// parse text
	zapLevel, err := zapcore.ParseLevel(level)
	if err == nil {
		return &zapLevel
	}

	// parse numeric
	levelNum, err := strconv.ParseInt(level, 10, 31)
	if err == nil {
		zapLevel = zapcore.Level(levelNum)
		if zapLevel < zapcore.DebugLevel {
			zapLevel = zapcore.DebugLevel
		}
		if zapLevel >= zapcore.InvalidLevel {
			zapLevel = zapcore.InvalidLevel - 1
		}
		return &zapLevel
	}

	// parse failed
	return nil
}
