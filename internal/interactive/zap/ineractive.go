package zap

import (
	"github.com/masinger/incredible/internal/interactive"
	"go.uber.org/zap"
)

type zapInteractive struct {
	logger *zap.Logger
}

func (z *zapInteractive) StartAction(message string) interactive.Action {
	res := &zapAction{
		logger:  z.logger,
		message: message,
	}
	res.print()
	return res
}

func NewZapInteractive(logger *zap.Logger) interactive.Interactive {
	return &zapInteractive{
		logger: logger,
	}
}
