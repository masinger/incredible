package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
)

type Cli struct {
	Executable     string
	Ctx            context.Context
	AdditionalEnvs []string
}

func (c Cli) WithEnv(name string, value string) Cli {
	return Cli{
		Executable:     c.Executable,
		Ctx:            c.Ctx,
		AdditionalEnvs: append(c.AdditionalEnvs, fmt.Sprintf("%s=%s", name, value)),
	}
}

func (c Cli) newCommand(args ...string) *exec.Cmd {
	var cmd *exec.Cmd
	if c.Ctx != nil {
		cmd = exec.Command(c.Executable, args...)
	} else {
		cmd = exec.CommandContext(c.Ctx, c.Executable, args...)
	}

	if c.AdditionalEnvs != nil && len(c.AdditionalEnvs) > 0 {
		cmd.Env = cmd.Environ()
		cmd.Env = append(cmd.Env, c.AdditionalEnvs...)
	}

	return cmd
}

func (c Cli) RunInheritStreams(args ...string) error {
	cmd := c.newCommand(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (c Cli) RunReadStdout(args ...string) (string, error) {
	cmd := c.newCommand(args...)
	buff := bytes.Buffer{}
	cmd.Stdout = &buff
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	err := cmd.Run()
	return buff.String(), err
}

func (c Cli) RunRedirectStdout(args ...string) (io.Reader, error) {
	cmd := c.newCommand(args...)
	reader, writer := io.Pipe()
	cmd.Stdout = writer
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	startErr := cmd.Start()
	if startErr != nil {
		return nil, startErr
	}

	go func() {
		err := cmd.Wait()
		if err != nil {
			_ = writer.CloseWithError(err)
		} else {
			_ = writer.Close()
		}
	}()

	return reader, nil
}
