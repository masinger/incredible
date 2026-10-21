package provider

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"
	"time"

	"github.com/masinger/incredible/internal/testutil/fakecli"
	bitwarden "github.com/masinger/incredible/pkg/bitwarden/specs"
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

// initFakeBw installs a fake bw following rules and returns an initialized
// provider. Its timeout leaves room for starting the fake on slow machines.
func initFakeBw(t *testing.T, rules ...fakecli.Rule) *Provider {
	t.Helper()
	fakecli.Install(t, "bw", fakecli.Spec{
		Rules:   rules,
		Default: fakecli.Behavior{Stderr: "fake bw: unexpected arguments", ExitCode: 99},
	})
	p := &Provider{}
	runtime := &provider.Runtime{Log: zap.NewNop().Sugar(), Timeout: fakecli.Timeout(t, "bw")}
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

func TestPrepareUsageTimesOutStatusChecks(t *testing.T) {
	tests := []struct {
		name    string
		rules   []fakecli.Rule
		command string
	}{
		{
			name:    "login",
			rules:   []fakecli.Rule{rule(hang, "login", "--check"), rule(succeed, "login", "--raw")},
			command: "bw login --check",
		},
		{
			name: "unlock",
			rules: []fakecli.Rule{
				rule(succeed, "login", "--check"),
				rule(hang, "unlock", "--check"),
				rule(succeed, "unlock", "--raw"),
			},
			command: "bw unlock --check",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := initFakeBw(t, tt.rules...)

			// A timed out check must fail instead of starting a login.
			err := fakecli.Await(t, func() error { return p.PrepareUsage(context.Background()) })
			requireTimeout(t, err, tt.command)
		})
	}
}

func TestPrepareUsageDoesNotTimeOutPrompts(t *testing.T) {
	tests := []struct {
		name  string
		rules []fakecli.Rule
	}{
		{
			name:  "login",
			rules: []fakecli.Rule{rule(fail, "login", "--check"), rule(hang, "login", "--raw")},
		},
		{
			name: "unlock",
			rules: []fakecli.Rule{
				rule(succeed, "login", "--check"),
				rule(fail, "unlock", "--check"),
				rule(hang, "unlock", "--raw"),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := initFakeBw(t, tt.rules...)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			// Cancel well after the timeout would have killed the prompt.
			time.AfterFunc(2*p.runtime.Timeout, cancel)

			err := fakecli.Await(t, func() error { return p.PrepareUsage(ctx) })
			if !errors.Is(err, context.Canceled) {
				t.Errorf("err = %v, want context.Canceled", err)
			}
		})
	}
}

func TestSourcesTimeOut(t *testing.T) {
	p := initFakeBw(t, rule(hang, "get", "item"), rule(hang, "get", "attachment"))
	ctx := context.Background()
	attachment := "attachment-name"

	t.Run("value", func(t *testing.T) {
		src, err := p.Open(ctx, &specs.Source{Bitwarden: &bitwarden.Source{Entry: "item-id"}})
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		err = fakecli.Await(t, func() error {
			_, err := src.(source.ValueSource).Get(ctx)
			return err
		})
		requireTimeout(t, err, "bw get item")
	})

	t.Run("attachment", func(t *testing.T) {
		src, err := p.Open(ctx, &specs.Source{Bitwarden: &bitwarden.Source{Entry: "item-id", Attachment: &attachment}})
		if err != nil {
			t.Fatalf("Open: %v", err)
		}
		err = fakecli.Await(t, func() error {
			reader, err := src.(source.BinarySource).Read(ctx)
			if err != nil {
				return err
			}
			_, err = io.ReadAll(reader)
			return err
		})
		requireTimeout(t, err, "bw get attachment")
	})
}
