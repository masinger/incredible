package provider

import (
	"context"
	"github.com/masinger/incredible/pkg/provider/source"
	"github.com/masinger/incredible/pkg/specs"
)

var Providers []Provider

type Provider interface {
	Name() string
	SupportsSource(src *specs.Source) bool
	Initialize(ctx context.Context, runtime *Runtime) error
	PrepareUsage(ctx context.Context) error
	Open(ctx context.Context, src *specs.Source) source.Source
}
