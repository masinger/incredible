package execution

import (
	"context"
	"go.uber.org/zap"
	"time"
)

type Options struct {
	Context                  context.Context
	Log                      *zap.SugaredLogger
	FailUnavailableProviders bool
	WarnUnavailableProviders bool
	// ProviderTimeout bounds non-interactive provider calls. Zero means no
	// timeout.
	ProviderTimeout time.Duration
}
