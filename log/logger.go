package log

import (
	"os"
	"strconv"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type LogConfig struct {
	JSONFormat bool
	LogLevel   string
}

var (
	once   sync.Once
	Logger *zap.SugaredLogger
)

func InitLogger(config LogConfig) {
	once.Do(func() {
		stdout := zapcore.AddSync(os.Stdout)

		cfg := zap.NewProductionEncoderConfig()
		cfg.TimeKey = "timestamp"
		cfg.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncodeLevel = zapcore.CapitalLevelEncoder

		var encoder zapcore.Encoder
		encoder = zapcore.NewConsoleEncoder(cfg)
		if config.JSONFormat {
			encoder = zapcore.NewJSONEncoder(cfg)
		}
		core := zapcore.NewCore(encoder, stdout, parseLevel(config.LogLevel))

		Logger = zap.New(core).Sugar()
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
