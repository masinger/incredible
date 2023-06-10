package zap

import (
	"fmt"
	"github.com/masinger/incredible/internal/interactive"
	"go.uber.org/zap"
)

type zapInteractive struct {
	logger *zap.Logger
}

func (z *zapInteractive) Confirm(confirmation interactive.Confirmation) (bool, error) {
	if confirmation.DisableDefault {
		return false, fmt.Errorf("non interactive session")
	}
	return confirmation.Default, nil
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
