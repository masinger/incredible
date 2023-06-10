package provider

import (
	"context"
	"github.com/masinger/incredible/pkg/cli"
	"github.com/masinger/incredible/pkg/provider"
	"github.com/masinger/incredible/pkg/provider/source"
	"github.com/masinger/incredible/pkg/specs"
	"os/exec"
)

type Provider struct {
	runtime *provider.Runtime
}

func (p *Provider) client(ctx context.Context) *CliClient {
	return &CliClient{
		cli.Cli{
			Ctx:        ctx,
			Executable: "az",
		},
	}
}

func (p *Provider) Name() string {
	return "azure"
}

func (p *Provider) SupportsSource(src *specs.Source) bool {
	return src.AzureKeyVaultSecret != nil
}

func (p *Provider) Initialize(ctx context.Context, runtime *provider.Runtime) error {
	_, err := exec.LookPath("az")
	if err != nil {
		return provider.NewProviderUnavailableErr(err)
	}
	return nil
}

func (p *Provider) PrepareUsage(ctx context.Context) error {
	client := p.client(ctx)
	return client.login()
}

func (p *Provider) Open(ctx context.Context, src *specs.Source) source.Source {
	keyVaultSecretSource := src.AzureKeyVaultSecret

	return &secretSource{
		itemId:   keyVaultSecretSource.ItemId,
		provider: p,
	}
}

func NewProvider() provider.Provider {
	return &Provider{}
}

func init() {
	provider.Providers = append(provider.Providers, NewProvider())
}
