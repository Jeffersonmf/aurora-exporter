package util

import (
	"go.uber.org/zap"
)

var Sugar *zap.SugaredLogger

func init() {
	logger, err := zap.NewProduction()
	if err != nil {
		Sugar.Infof("can't initialize zap logger: %v", err)
	}

	Sugar = logger.Sugar()
	if err != nil {
		Sugar.Infof(err.Error())
	}
}
