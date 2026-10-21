// Package fakecli replaces external provider CLIs (bw, az, ...) in tests.
//
// It uses the helper-process pattern: the test binary itself plays the fake
// CLI, so no shell scripts are needed and tests run on Linux and Windows
// alike. The fake's behaviour is described by a Spec, which is passed to the
// child process as JSON in an environment variable.
//
// Every test package that uses fakecli needs a TestMain calling Dispatch
// first:
//
//	func TestMain(m *testing.M) {
//		fakecli.Dispatch()
//		os.Exit(m.Run())
//	}
//
// Use Executable to point a cli.Cli at the fake:
//
//	path, env := fakecli.Executable(t, fakecli.Spec{
//		Default: fakecli.Behavior{Stdout: "hello"},
//	})
//	c := cli.Cli{Ctx: ctx, Executable: path, AdditionalEnvs: env}
//
// Use Install when the code under test runs a CLI by name (e.g. "bw"):
//
//	fakecli.Install(t, "bw", fakecli.Spec{
//		Rules: []fakecli.Rule{
//			{Args: []string{"unlock", "--check"}, Behavior: fakecli.Behavior{ExitCode: 1}},
//			{Args: []string{"unlock", "--raw"}, Behavior: fakecli.Behavior{Stdout: "session"}},
//		},
//	})
//
// Install changes the process environment, so tests using it must not call
// t.Parallel. A fake that hangs must be killed and waited for before the test
// ends; on Windows its executable can't be removed while it is running. Run
// code that should kill a hanging fake through Await, so the test fails fast
// if it doesn't.
package fakecli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"
)

const specEnv = "INCREDIBLE_FAKECLI_SPEC"

const hangDuration = 10 * time.Minute

// awaitTimeout is how long Await waits. It is far shorter than hangDuration and
// the default timeout of go test.
const awaitTimeout = 30 * time.Second

// Behavior describes what the fake does. The steps run in field order.
type Behavior struct {
	// EchoStdin copies stdin to stdout until EOF.
	EchoStdin bool `json:",omitempty"`
	// Stdout is written to stdout. Keep it small, Windows limits the size of
	// environment variables. Use StdoutBytes for large output.
	Stdout string `json:",omitempty"`
	// StdoutBytes is the number of additional bytes ('x') written to stdout.
	StdoutBytes int    `json:",omitempty"`
	Stderr      string `json:",omitempty"`
	// LeaveChild starts a copy of the fake that inherits stdout, sleeps for
	// this duration and isn't waited for. It acts like a wrapper script that
	// leaves a process behind. Don't combine it with Install on Windows: the
	// test's temp dir can't be removed while the child runs.
	LeaveChild time.Duration `json:",omitempty"`
	Sleep      time.Duration `json:",omitempty"`
	// Hang blocks until the process is killed (at most 10 minutes).
	Hang     bool `json:",omitempty"`
	ExitCode int  `json:",omitempty"`
}

// Rule applies Behavior if the CLI arguments start with Args.
type Rule struct {
	Args []string
	Behavior
}

// Spec describes the fake CLI. The first matching Rule wins; Default applies
// if none matches.
type Spec struct {
	Rules   []Rule
	Default Behavior
}

func (s Spec) behaviorFor(args []string) Behavior {
	for _, rule := range s.Rules {
		if len(args) >= len(rule.Args) && slices.Equal(args[:len(rule.Args)], rule.Args) {
			return rule.Behavior
		}
	}
	return s.Default
}

var dispatched bool

// Dispatch runs the fake and exits if the test binary has been started as a
// fake CLI by Executable or Install. Otherwise, it returns immediately.
// Call it first in TestMain.
func Dispatch() {
	dispatched = true
	encoded, ok := os.LookupEnv(specEnv)
	if !ok {
		return
	}
	var spec Spec
	if err := json.Unmarshal([]byte(encoded), &spec); err != nil {
		fmt.Fprintf(os.Stderr, "fakecli: invalid spec: %v\n", err)
		os.Exit(127)
	}
	os.Exit(run(spec.behaviorFor(os.Args[1:])))
}

func run(b Behavior) int {
	if err := write(b); err != nil {
		fmt.Fprintf(os.Stderr, "fakecli: %v\n", err)
		return 126
	}
	if b.LeaveChild > 0 {
		if err := leaveChild(b.LeaveChild); err != nil {
			fmt.Fprintf(os.Stderr, "fakecli: %v\n", err)
			return 126
		}
	}
	time.Sleep(b.Sleep)
	if b.Hang {
		time.Sleep(hangDuration)
	}
	return b.ExitCode
}

func leaveChild(sleep time.Duration) error {
	self, err := os.Executable()
	if err != nil {
		return err
	}
	encoded, err := json.Marshal(Spec{Default: Behavior{Sleep: sleep}})
	if err != nil {
		return err
	}
	child := exec.Command(self)
	child.Env = append(os.Environ(), specEnv+"="+string(encoded))
	child.Stdout = os.Stdout
	return child.Start()
}

func write(b Behavior) error {
	if b.EchoStdin {
		if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
			return err
		}
	}
	if _, err := os.Stdout.WriteString(b.Stdout); err != nil {
		return err
	}
	chunk := bytes.Repeat([]byte{'x'}, 32*1024)
	for remaining := b.StdoutBytes; remaining > 0; remaining -= len(chunk) {
		if _, err := os.Stdout.Write(chunk[:min(remaining, len(chunk))]); err != nil {
			return err
		}
	}
	_, err := os.Stderr.WriteString(b.Stderr)
	return err
}

// Executable returns the path of the fake CLI and the environment entries
// selecting spec. Pass them to cli.Cli as Executable and AdditionalEnvs.
func Executable(t testing.TB, spec Spec) (path string, env []string) {
	t.Helper()
	requireDispatch(t)
	path, err := os.Executable()
	if err != nil {
		t.Fatalf("fakecli: %v", err)
	}
	return path, []string{specEnv + "=" + encode(t, spec)}
}

// Install puts a fake CLI called name on PATH for the rest of the test, so
// running name or exec.LookPath(name) finds it.
func Install(t testing.TB, name string, spec Spec) {
	t.Helper()
	requireDispatch(t)
	self, err := os.Executable()
	if err != nil {
		t.Fatalf("fakecli: %v", err)
	}
	dir := t.TempDir()
	target := filepath.Join(dir, name)
	if runtime.GOOS == "windows" {
		target += ".exe"
	}
	if err := copyExecutable(self, target); err != nil {
		t.Fatalf("fakecli: %v", err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv(specEnv, encode(t, spec))
}

// Timeout returns a timeout for code that runs the fake installed as name a
// few times: at least a second, or five times as long as the fake took to start
// and exit. Starting can be slow, e.g. due to virus scanning on Windows or the
// race detector.
func Timeout(t testing.TB, name string) time.Duration {
	t.Helper()
	cmd := exec.Command(name)
	// Override the installed spec, so the fake exits right away.
	cmd.Env = append(os.Environ(), specEnv+"={}")
	start := time.Now()
	if err := cmd.Run(); err != nil {
		t.Fatalf("fakecli: running %s: %v", name, err)
	}
	return max(time.Second, 5*time.Since(start))
}

// Await runs fn and returns its error. It fails the test if fn doesn't return
// within 30 seconds, e.g. because a hanging fake hasn't been killed.
func Await(t testing.TB, fn func() error) error {
	t.Helper()
	done := make(chan error, 1)
	go func() {
		done <- fn()
	}()
	select {
	case err := <-done:
		return err
	case <-time.After(awaitTimeout):
		t.Fatalf("fakecli: still running after %v", awaitTimeout)
		return nil
	}
}

func requireDispatch(t testing.TB) {
	t.Helper()
	if !dispatched {
		// Without Dispatch, the fake would run the package's tests again.
		t.Fatal("fakecli: call fakecli.Dispatch() first in TestMain")
	}
}

func encode(t testing.TB, spec Spec) string {
	t.Helper()
	encoded, err := json.Marshal(spec)
	if err != nil {
		t.Fatalf("fakecli: %v", err)
	}
	return string(encoded)
}

func copyExecutable(src string, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() {
		_ = in.Close()
	}()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o755)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()
	_, err = io.Copy(out, in)
	return err
}
