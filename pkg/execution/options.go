package execution

import (
	"context"
	"go.uber.org/zap"
)

type Options struct {
	Context                  context.Context
	Log                      *zap.SugaredLogger
	FailUnavailableProviders bool
	WarnUnavailableProviders bool
}
