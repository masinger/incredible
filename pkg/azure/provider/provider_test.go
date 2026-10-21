package provider

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"github.com/masinger/incredible/internal/testutil/fakecli"
	azure "github.com/masinger/incredible/pkg/azure/specs"
	"github.com/masinger/incredible/pkg/cli"
	"github.com/masinger/incredible/pkg/provider"
	"github.com/masinger/incredible/pkg/provider/source"
	"github.com/masinger/incredible/pkg/specs"
	"go.uber.org/zap"
)

func TestMain(m *testing.M) {
	fakecli.Dispatch()
	os.Exit(m.Run())
}

var (
	succeed = fakecli.Behavior{}
	fail    = fakecli.Behavior{ExitCode: 1}
	hang    = fakecli.Behavior{Hang: true}
)

func rule(behavior fakecli.Behavior, args ...string) fakecli.Rule {
	return fakecli.Rule{Args: args, Behavior: behavior}
}

// initFakeAz installs a fake az following rules and returns an initialized
// provider. Its timeout leaves room for starting the fake on slow machines.
func initFakeAz(t *testing.T, rules ...fakecli.Rule) *Provider {
	t.Helper()
	fakecli.Install(t, "az", fakecli.Spec{
		Rules:   rules,
		Default: fakecli.Behavior{Stderr: "fake az: unexpected arguments", ExitCode: 99},
	})
	p := &Provider{}
	runtime := &provider.Runtime{Log: zap.NewNop().Sugar(), Timeout: fakecli.Timeout(t, "az")}
	if err := p.Initialize(context.Background(), runtime); err != nil {
		t.Fatalf("Initialize: %v", err)
	}
	return p
}

func requireTimeout(t *testing.T, err error, command string) {
	t.Helper()
	var timeoutErr *cli.TimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("err = %v, want *cli.TimeoutError", err)
	}
	if timeoutErr.Command != command {
		t.Errorf("Command = %q, want %q", timeoutErr.Command, command)
	}
}

func TestPrepareUsageTimesOutAccountCheck(t *testing.T) {
	p := initFakeAz(t, rule(hang, "account", "list"), rule(succeed, "login"))

	// A timed out check must fail instead of starting a login.
	err := fakecli.Await(t, func() error { return p.PrepareUsage(context.Background()) })
	requireTimeout(t, err, "az account list")
}

func TestPrepareUsageDoesNotTimeOutLogin(t *testing.T) {
	p := initFakeAz(t, rule(fail, "account", "list"), rule(hang, "login"))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Cancel well after the timeout would have killed the login.
	time.AfterFunc(2*p.runtime.Timeout, cancel)

	err := fakecli.Await(t, func() error { return p.PrepareUsage(ctx) })
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestSecretSourceTimesOut(t *testing.T) {
	p := initFakeAz(t, rule(hang, "keyvault", "secret", "show"))
	ctx := context.Background()
	src, err := p.Open(ctx, &specs.Source{AzureKeyVaultSecret: &azure.KeyVaultSecretSource{ItemId: "secret-id"}})
	if err != nil {
		t.Fatalf("Open: %v", err)
	}

	err = fakecli.Await(t, func() error {
		_, err := src.(source.ValueSource).Get(ctx)
		return err
	})
	requireTimeout(t, err, "az keyvault secret")
}

func TestSecretSourceChecksExitStatus(t *testing.T) {
	const output = `{"value": "secret"}`
	tests := []struct {
		name     string
		exitCode int
		want     string
		wantErr  bool
	}{
		{name: "success", exitCode: 0, want: "secret"},
		// az has printed the secret, but failed afterwards.
		{name: "failure", exitCode: 1, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := initFakeAz(t, rule(fakecli.Behavior{Stdout: output, ExitCode: tt.exitCode}, "keyvault", "secret", "show"))
			ctx := context.Background()
			src, err := p.Open(ctx, &specs.Source{AzureKeyVaultSecret: &azure.KeyVaultSecretSource{ItemId: "secret-id"}})
			if err != nil {
				t.Fatalf("Open: %v", err)
			}

			got, err := src.(source.ValueSource).Get(ctx)
			if (err != nil) != tt.wantErr || got != tt.want {
				t.Errorf("Get = %q, %v, want %q, error: %v", got, err, tt.want, tt.wantErr)
			}
		})
	}
}
