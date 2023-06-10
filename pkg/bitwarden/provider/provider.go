package provider

import (
	"context"
	bitwardenSpecs "github.com/masinger/incredible/pkg/bitwarden/specs"
	"github.com/masinger/incredible/pkg/cli"
	"github.com/masinger/incredible/pkg/provider"
	"github.com/masinger/incredible/pkg/provider/source"
	"github.com/masinger/incredible/pkg/specs"
	"go.uber.org/zap"
	"os/exec"
)

type Provider struct {
	sessionKey string
	runtime    *provider.Runtime
}

func (p *Provider) Name() string {
	return "bitwarden"
}

func (p *Provider) Initialize(ctx context.Context, runtime *provider.Runtime) error {
	_, err := exec.LookPath("bw")
	if err != nil {
		return provider.NewProviderUnavailableErr(err)
	}
	p.runtime = runtime
	return nil
}

func (p *Provider) SupportsSource(src *specs.Source) bool {
	return src.Bitwarden != nil
}

func (p *Provider) PrepareUsage(ctx context.Context) error {
	b := CliClient{
		cli.Cli{
			Ctx:        ctx,
			Executable: "bw",
		},
	}
	sessionKey, err := b.activeSession()
	if err != nil {
		return err
	}
	p.sessionKey = sessionKey
	p.runtime.Log.Debugw("Obtained session", zap.String("session", sessionKey))
	return nil
}

func (p *Provider) Open(ctx context.Context, src *specs.Source) source.Source {
	bwSrc := src.Bitwarden
	if bwSrc.Attachment != nil {
		return &fileSource{
			sessionKey:     p.sessionKey,
			itemId:         bwSrc.Entry,
			attachmentName: *bwSrc.Attachment,
		}
	} else {
		var accessor fieldAccessor
		var field = bitwardenSpecs.FieldPassword
		if bwSrc.Field != nil {
			field = *bwSrc.Field
		}

		switch field {
		case bitwardenSpecs.FieldUsername:
			accessor = FieldAccessorUsername
			break
		case bitwardenSpecs.FieldPassword:
			fallthrough
		default:
			accessor = FieldAccessorPassword
		}
		var valSource source.ValueSource = &valueSource{
			sessionKey:    p.sessionKey,
			itemId:        bwSrc.Entry,
			fieldAccessor: accessor,
		}
		return valSource
	}
}

func NewProvider() provider.Provider {
	return &Provider{}
}

func init() {
	provider.Providers = append(provider.Providers, NewProvider())
}
