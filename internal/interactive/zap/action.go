package zap

import (
	"go.uber.org/zap"
)

type zapAction struct {
	logger  *zap.Logger
	message string
}

func (action zapAction) print() {
	action.logger.Info(action.message, zap.String("status", "executing"))
}

func (action zapAction) Complete(message string) {
	if message != "" {
		action.logger.Info(message, zap.String("status", "complete"))
	} else {
		action.logger.Info(action.message, zap.String("status", "complete"))
	}
}
