package bootstrap

import (
	"go.uber.org/zap"
)

func registerEventHandlers(infra *infrastructure, log *zap.Logger) {
	log.Info("registering authorization event handlers")

	registerNotificationHandlers(infra, log)
}
