package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// waitDelay bounds how long a command's I/O pipes are waited for after it has
// been killed or has exited. Wrappers like the az script or bw.cmd don't pass
// the kill on to the process they start, which keeps the pipes open.
var waitDelay = 5 * time.Second

type Cli struct {
	Executable     string
	Ctx            context.Context
	AdditionalEnvs []string
	// Timeout bounds the runtime of each command. Zero means no timeout.
	Timeout time.Duration
}

// TimeoutError is returned when a command has been killed because its
// Cli.Timeout elapsed.
type TimeoutError struct {
	// Command is the executable followed by at most two leading arguments,
	// e.g. "bw get item". Later arguments, like IDs, are left out.
	Command string
	Timeout time.Duration
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("%s timed out after %s", e.Command, e.Timeout)
}

func (e *TimeoutError) Unwrap() error {
	return context.DeadlineExceeded
}

// Succeeded reports whether a command that returned err exited with status
// zero. A non-zero exit status yields false and no error. Any other error,
// e.g. a TimeoutError, is returned.
func Succeeded(err error) (bool, error) {
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false, nil
	}
	return err == nil, err
}

func (c Cli) WithEnv(name string, value string) Cli {
	return Cli{
		Executable:     c.Executable,
		Ctx:            c.Ctx,
		AdditionalEnvs: append(c.AdditionalEnvs, fmt.Sprintf("%s=%s", name, value)),
		Timeout:        c.Timeout,
	}
}

// WithoutTimeout returns a copy of c without a timeout. Use it for commands
// waiting for user input, like logins.
func (c Cli) WithoutTimeout() Cli {
	c.Timeout = 0
	return c
}

// newCommand returns the command and a function that must be called with the
// result of running it. That function releases the command's context and
// replaces the error with the cause if the context is done.
func (c Cli) newCommand(args ...string) (*exec.Cmd, func(error) error) {
	ctx := c.Ctx
	if ctx == nil {
		ctx = context.Background()
	}
	cancel := context.CancelFunc(func() {})
	if c.Timeout > 0 {
		ctx, cancel = context.WithTimeoutCause(ctx, c.Timeout, &TimeoutError{
			Command: commandName(c.Executable, args),
			Timeout: c.Timeout,
		})
	}

	cmd := exec.CommandContext(ctx, c.Executable, args...)
	cmd.WaitDelay = waitDelay

	if c.AdditionalEnvs != nil && len(c.AdditionalEnvs) > 0 {
		cmd.Env = cmd.Environ()
		cmd.Env = append(cmd.Env, c.AdditionalEnvs...)
	}

	finish := func(err error) error {
		defer cancel()
		if err != nil && ctx.Err() != nil {
			return context.Cause(ctx)
		}
		return err
	}
	return cmd, finish
}

func commandName(executable string, args []string) string {
	parts := append([]string{filepath.Base(executable)}, args[:min(len(args), 2)]...)
	return strings.Join(parts, " ")
}

func (c Cli) RunInheritStreams(args ...string) error {
	cmd, finish := c.newCommand(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return finish(cmd.Run())
}

func (c Cli) RunReadStdout(args ...string) (string, error) {
	cmd, finish := c.newCommand(args...)
	buff := bytes.Buffer{}
	cmd.Stdout = &buff
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	if errors.Is(err, exec.ErrWaitDelay) {
		// The command exited successfully, but left a process behind that
		// still holds stdout. All output of the command has been read.
		err = nil
	}
	return buff.String(), finish(err)
}

func (c Cli) RunRedirectStdout(args ...string) (io.Reader, error) {
	cmd, finish := c.newCommand(args...)
	reader, writer := io.Pipe()
	cmd.Stdout = writer
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	startErr := cmd.Start()
	if startErr != nil {
		return nil, finish(startErr)
	}

	go func() {
		// The context must stay alive until the command has finished.
		err := finish(cmd.Wait())
		if err != nil {
			_ = writer.CloseWithError(err)
		} else {
			_ = writer.Close()
		}
	}()

	return reader, nil
}
