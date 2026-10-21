package cli_test

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/masinger/incredible/internal/testutil/fakecli"
	"github.com/masinger/incredible/pkg/cli"
)

func TestMain(m *testing.M) {
	fakecli.Dispatch()
	os.Exit(m.Run())
}

func fakeCli(t *testing.T, spec fakecli.Spec) cli.Cli {
	t.Helper()
	path, env := fakecli.Executable(t, spec)
	// Ctx is left nil, which must work as well.
	return cli.Cli{
		Executable:     path,
		AdditionalEnvs: env,
	}
}

// runners run a command with each of the Run methods.
var runners = map[string]func(c cli.Cli, args ...string) error{
	"RunInheritStreams": func(c cli.Cli, args ...string) error {
		return c.RunInheritStreams(args...)
	},
	"RunReadStdout": func(c cli.Cli, args ...string) error {
		_, err := c.RunReadStdout(args...)
		return err
	},
	"RunRedirectStdout": func(c cli.Cli, args ...string) error {
		reader, err := c.RunRedirectStdout(args...)
		if err != nil {
			return err
		}
		_, err = io.ReadAll(reader)
		return err
	},
}

func TestRunReadStdoutReturnsOutput(t *testing.T) {
	c := fakeCli(t, fakecli.Spec{Default: fakecli.Behavior{Stdout: "hello"}})

	out, err := c.RunReadStdout("get", "item")
	if err != nil {
		t.Fatalf("RunReadStdout: %v", err)
	}
	if out != "hello" {
		t.Errorf("output = %q, want %q", out, "hello")
	}
}

func TestRunReadStdoutReturnsExitError(t *testing.T) {
	c := fakeCli(t, fakecli.Spec{Default: fakecli.Behavior{ExitCode: 3}})

	_, err := c.RunReadStdout("get", "item")
	var exitErr *exec.ExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("err = %v, want *exec.ExitError", err)
	}
	if exitErr.ExitCode() != 3 {
		t.Errorf("exit code = %d, want 3", exitErr.ExitCode())
	}
}

func TestRunRedirectStdoutStreamsAllOutput(t *testing.T) {
	const size = 1 << 20
	c := fakeCli(t, fakecli.Spec{Default: fakecli.Behavior{StdoutBytes: size}})

	reader, err := c.RunRedirectStdout("get", "attachment")
	if err != nil {
		t.Fatalf("RunRedirectStdout: %v", err)
	}
	out, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if len(out) != size {
		t.Errorf("read %d bytes, want %d", len(out), size)
	}
}

func TestRunReadStdoutIgnoresLeftOverProcess(t *testing.T) {
	cli.SetWaitDelay(t, 100*time.Millisecond)
	c := fakeCli(t, fakecli.Spec{Default: fakecli.Behavior{Stdout: "hello", LeaveChild: 2 * time.Second}})

	out, err := c.RunReadStdout("account", "list")
	if err != nil {
		t.Fatalf("RunReadStdout: %v", err)
	}
	if out != "hello" {
		t.Errorf("output = %q, want %q", out, "hello")
	}
}

func TestTimeoutDoesNotWaitForLeftOverProcess(t *testing.T) {
	cli.SetWaitDelay(t, 100*time.Millisecond)
	c := fakeCli(t, fakecli.Spec{Default: fakecli.Behavior{LeaveChild: 3 * time.Second, Hang: true}})
	c.Timeout = 200 * time.Millisecond

	start := time.Now()
	_, err := c.RunReadStdout("get", "item")
	var timeoutErr *cli.TimeoutError
	if !errors.As(err, &timeoutErr) {
		t.Fatalf("err = %v, want *cli.TimeoutError", err)
	}
	if elapsed := time.Since(start); elapsed > 2500*time.Millisecond {
		t.Errorf("took %v, the left over process has been waited for", elapsed)
	}
}

func TestCancelKillsCommand(t *testing.T) {
	for name, run := range runners {
		t.Run(name, func(t *testing.T) {
			c := fakeCli(t, fakecli.Spec{Default: fakecli.Behavior{Hang: true}})
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			c.Ctx = ctx
			time.AfterFunc(200*time.Millisecond, cancel)

			err := fakecli.Await(t, func() error { return run(c, "get", "item") })
			if !errors.Is(err, context.Canceled) {
				t.Errorf("err = %v, want context.Canceled", err)
			}
		})
	}
}

func TestTimeoutKillsCommand(t *testing.T) {
	for name, run := range runners {
		t.Run(name, func(t *testing.T) {
			c := fakeCli(t, fakecli.Spec{Default: fakecli.Behavior{Hang: true}})
			c.Timeout = 200 * time.Millisecond

			err := fakecli.Await(t, func() error { return run(c, "get", "item", "item-id", "--raw") })
			var timeoutErr *cli.TimeoutError
			if !errors.As(err, &timeoutErr) {
				t.Fatalf("err = %v, want *cli.TimeoutError", err)
			}
			if want := filepath.Base(c.Executable) + " get item"; timeoutErr.Command != want {
				t.Errorf("Command = %q, want %q", timeoutErr.Command, want)
			}
			if strings.Contains(err.Error(), "item-id") {
				t.Errorf("error %q contains a later argument", err)
			}
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Errorf("err = %v, want it to wrap context.DeadlineExceeded", err)
			}
		})
	}
}

func TestWithEnvKeepsTimeout(t *testing.T) {
	c := cli.Cli{Timeout: time.Minute}.WithEnv("NAME", "value")
	if c.Timeout != time.Minute {
		t.Errorf("Timeout = %v, want %v", c.Timeout, time.Minute)
	}
}

func TestWithoutTimeoutDoesNotKillCommand(t *testing.T) {
	c := fakeCli(t, fakecli.Spec{Default: fakecli.Behavior{Hang: true}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	c.Ctx = ctx
	c.Timeout = 100 * time.Millisecond
	// Cancel well after the timeout would have killed the command.
	time.AfterFunc(time.Second, cancel)

	err := fakecli.Await(t, func() error { return c.WithoutTimeout().RunInheritStreams("login") })
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}

func TestSucceeded(t *testing.T) {
	exitErr := fakeCli(t, fakecli.Spec{Default: fakecli.Behavior{ExitCode: 1}}).RunInheritStreams()
	timeoutErr := &cli.TimeoutError{Command: "bw login --check", Timeout: time.Minute}
	tests := []struct {
		name    string
		err     error
		want    bool
		wantErr error
	}{
		{name: "exit status zero", err: nil, want: true},
		{name: "non-zero exit status", err: exitErr, want: false},
		{name: "cancelled", err: context.Canceled, want: false, wantErr: context.Canceled},
		{name: "timeout", err: timeoutErr, want: false, wantErr: timeoutErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cli.Succeeded(tt.err)
			if got != tt.want || err != tt.wantErr {
				t.Errorf("Succeeded(%v) = %v, %v, want %v, %v", tt.err, got, err, tt.want, tt.wantErr)
			}
		})
	}
}
