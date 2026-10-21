package fakecli_test

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/masinger/incredible/internal/testutil/fakecli"
)

func TestMain(m *testing.M) {
	fakecli.Dispatch()
	os.Exit(m.Run())
}

type result struct {
	stdout   string
	stderr   string
	exitCode int
}

func runFake(t *testing.T, cmd *exec.Cmd, stdin string) result {
	t.Helper()
	var stdout, stderr bytes.Buffer
	cmd.Stdin = strings.NewReader(stdin)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	var exitErr *exec.ExitError
	if err != nil && !errors.As(err, &exitErr) {
		t.Fatalf("running fake: %v", err)
	}
	return result{stdout: stdout.String(), stderr: stderr.String(), exitCode: cmd.ProcessState.ExitCode()}
}

func command(t *testing.T, ctx context.Context, spec fakecli.Spec, args ...string) *exec.Cmd {
	t.Helper()
	path, env := fakecli.Executable(t, spec)
	cmd := exec.CommandContext(ctx, path, args...)
	cmd.Env = append(os.Environ(), env...)
	return cmd
}

func TestRulesMatchArgumentPrefix(t *testing.T) {
	spec := fakecli.Spec{
		Rules: []fakecli.Rule{
			{Args: []string{"unlock", "--check"}, Behavior: fakecli.Behavior{ExitCode: 1}},
			{Args: []string{"unlock"}, Behavior: fakecli.Behavior{Stdout: "session"}},
		},
		Default: fakecli.Behavior{Stdout: "default"},
	}
	tests := []struct {
		args []string
		want result
	}{
		{args: []string{"unlock", "--check"}, want: result{exitCode: 1}},
		{args: []string{"unlock", "--raw"}, want: result{stdout: "session"}},
		{args: []string{"get", "item", "unlock"}, want: result{stdout: "default"}},
		{args: nil, want: result{stdout: "default"}},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			ctx := context.Background()
			got := runFake(t, command(t, ctx, spec, tt.args...), "")
			if got != tt.want {
				t.Errorf("got %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestBehavior(t *testing.T) {
	ctx := context.Background()
	spec := fakecli.Spec{Default: fakecli.Behavior{
		EchoStdin:   true,
		Stdout:      "-out-",
		StdoutBytes: 100_000,
		Stderr:      "err",
		ExitCode:    3,
	}}
	got := runFake(t, command(t, ctx, spec), "in")

	wantStdout := "in-out-" + strings.Repeat("x", 100_000)
	if got.stdout != wantStdout {
		t.Errorf("stdout has %d bytes, want %d bytes", len(got.stdout), len(wantStdout))
	}
	if got.stderr != "err" {
		t.Errorf("stderr = %q, want %q", got.stderr, "err")
	}
	if got.exitCode != 3 {
		t.Errorf("exit code = %d, want 3", got.exitCode)
	}
}

func TestHangUntilKilled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	cmd := command(t, ctx, fakecli.Spec{Default: fakecli.Behavior{Hang: true}})

	start := time.Now()
	err := cmd.Run()
	if err == nil {
		t.Fatal("hanging fake exited successfully")
	}
	if elapsed := time.Since(start); elapsed > 30*time.Second {
		t.Errorf("killing the fake took %v", elapsed)
	}
}

func TestInstallPutsFakeOnPath(t *testing.T) {
	const name = "incredible-fakecli-test"
	fakecli.Install(t, name, fakecli.Spec{
		Rules: []fakecli.Rule{
			{Args: []string{"status"}, Behavior: fakecli.Behavior{Stdout: "installed"}},
		},
	})

	if _, err := exec.LookPath(name); err != nil {
		t.Fatalf("LookPath(%q): %v", name, err)
	}
	ctx := context.Background()
	got := runFake(t, exec.CommandContext(ctx, name, "status"), "")
	if got != (result{stdout: "installed"}) {
		t.Errorf("got %+v", got)
	}
}
