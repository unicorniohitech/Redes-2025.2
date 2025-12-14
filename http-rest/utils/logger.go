package utils

import (
	"go.uber.org/zap"
)

var Logger *zap.Logger

func init() {
	var err error
	Logger, err = zap.NewDevelopment()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}
}

func GetLogger() *zap.Logger {
	return Logger
}

func SyncLogger() {
	Logger.Sync()
}
