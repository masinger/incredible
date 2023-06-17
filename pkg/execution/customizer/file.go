package customizer

import (
	"context"
	"errors"
	"github.com/masinger/incredible/pkg/provider/source"
	"io"
	"os"
	"os/exec"
)

func writeFile(
	ctx context.Context,
	src source.BinarySource,
	file *os.File,
) (err error) {
	input, err := src.Read(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := input.Close(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
	}()

	_, err = io.Copy(file, input)
	return err
}

func file(
	ctx context.Context,
	src source.BinarySource,
	target *os.File,
) Customizer {
	return func(cmd *exec.Cmd) (cleanup Cleanup, err error) {
		cleanup = func(cmd *exec.Cmd) error {
			return os.Remove(target.Name())
		}

		return cleanup, writeFile(ctx, src, target)
	}
}

func File(
	ctx context.Context,
	src source.BinarySource,
	path string,
) Customizer {
	return func(cmd *exec.Cmd) (cleanup Cleanup, err error) {
		f, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		defer func() {
			if closeErr := f.Close(); closeErr != nil {
				err = errors.Join(err, closeErr)
			}
		}()
		return file(ctx, src, f)(cmd)
	}
}

func TempFile(
	ctx context.Context,
	src source.BinarySource,
	nameTarget *string,
) Customizer {
	return func(cmd *exec.Cmd) (Cleanup, error) {
		tempFile, err := os.CreateTemp("", "")
		if err != nil {
			return nil, err
		}
		*nameTarget = tempFile.Name()
		return file(
			ctx,
			src,
			tempFile,
		)(cmd)
	}
}
