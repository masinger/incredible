package execution_test

import (
	"context"
	"testing"
	"time"

	"github.com/masinger/incredible/pkg/execution"
	"github.com/masinger/incredible/pkg/provider"
	"github.com/masinger/incredible/pkg/provider/source"
	"github.com/masinger/incredible/pkg/specs"
	"go.uber.org/zap"
)

// recordingProvider records the runtime it has been initialized with.
type recordingProvider struct {
	runtime *provider.Runtime
}

func (p *recordingProvider) Name() string {
	return "recording"
}

func (p *recordingProvider) SupportsSource(*specs.Source) bool {
	return false
}

func (p *recordingProvider) Initialize(_ context.Context, runtime *provider.Runtime) error {
	p.runtime = runtime
	return nil
}

func (p *recordingProvider) PrepareUsage(context.Context) error {
	return nil
}

func (p *recordingProvider) Open(context.Context, *specs.Source) (source.Source, error) {
	return nil, nil
}

func TestNewExecutionPassesProviderTimeout(t *testing.T) {
	p := &recordingProvider{}
	_, err := execution.NewExecution([]provider.Provider{p}, execution.Options{
		Log:             zap.NewNop().Sugar(),
		ProviderTimeout: 42 * time.Second,
	})
	if err != nil {
		t.Fatalf("NewExecution: %v", err)
	}
	if p.runtime.Timeout != 42*time.Second {
		t.Errorf("runtime timeout = %v, want %v", p.runtime.Timeout, 42*time.Second)
	}
}
