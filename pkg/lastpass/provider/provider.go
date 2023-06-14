package provider

import (
	"context"
	"github.com/ansd/lastpass-go"
	"github.com/masinger/incredible/pkg/provider"
	"github.com/masinger/incredible/pkg/provider/source"
	"github.com/masinger/incredible/pkg/specs"
)

type Provider struct {
	clientCreator ClientCreator
	client        *lastpass.Client
}

func (p *Provider) Name() string {
	return "lastpass"
}

func (p *Provider) SupportsSource(src *specs.Source) bool {
	return src.LastPass != nil
}

func (p *Provider) Initialize(_ context.Context, _ *provider.Runtime) error {
	// No initialization required
	return nil
}

func (p *Provider) PrepareUsage(ctx context.Context) error {
	cl, err := p.clientCreator(ctx)
	if err != nil {
		return err
	}
	p.client = cl
	return nil
}

func (p *Provider) Open(_ context.Context, src *specs.Source) (source.Source, error) {
	lastpassSource := src.LastPass
	mapping, err := fieldMapping.DefaultAccessor(lastpassSource.Field)
	if err != nil {
		return nil, err
	}
	return &accountSource{
		id:            lastpassSource.AccountId,
		provider:      p,
		fieldAccessor: mapping,
	}, nil
}

func NewProvider(
	clientCreator ClientCreator,
) provider.Provider {
	return &Provider{
		clientCreator: clientCreator,
	}
}

func init() {
	provider.Providers = append(provider.Providers, NewProvider(InteractiveClientCreator))
}
