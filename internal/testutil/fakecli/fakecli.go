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
// ends; on Windows its executable can't be removed while it is running.
package fakecli

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"testing"
	"time"
)

const specEnv = "INCREDIBLE_FAKECLI_SPEC"

const hangDuration = 10 * time.Minute

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
	if b.Hang {
		time.Sleep(hangDuration)
	}
	return b.ExitCode
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
