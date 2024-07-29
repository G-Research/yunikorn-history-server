package log

import (
	"context"

	"go.uber.org/zap"

	"github.com/G-Research/yunikorn-history-server/internal/log"
	testconfig "github.com/G-Research/yunikorn-history-server/test/config"
)

func GetTestLogger(ctx context.Context) (context.Context, *zap.SugaredLogger) {
	logger := log.Init(testconfig.GetTestLogConfig())
	return log.ToContext(ctx, logger), logger
}
