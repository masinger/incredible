package cli

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
)

type Cli struct {
	Executable string
	Ctx        context.Context
}

func (c Cli) newCommand(args ...string) *exec.Cmd {
	if c.Ctx != nil {
		return exec.Command(c.Executable, args...)
	} else {
		return exec.CommandContext(c.Ctx, c.Executable, args...)
	}
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
