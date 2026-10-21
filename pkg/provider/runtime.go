package provider

import (
	"go.uber.org/zap"
	"time"
)

type Runtime struct {
	Log *zap.SugaredLogger
	// Timeout bounds non-interactive provider calls. Zero means no timeout.
	Timeout time.Duration
}
