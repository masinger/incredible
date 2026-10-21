package cli_test

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"

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
	return cli.Cli{
		// Set explicitly: newCommand passes a nil Ctx to exec.CommandContext.
		Ctx:            context.Background(),
		Executable:     path,
		AdditionalEnvs: env,
	}
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
