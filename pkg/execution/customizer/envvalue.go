package customizer

import (
	"context"
	"fmt"
	"github.com/masinger/incredible/pkg/provider/source"
	"os/exec"
)

func LoadValue(
	ctx context.Context,
	source source.ValueSource,
	target *string,
) Customizer {
	return func(cmd *exec.Cmd) (Cleanup, error) {
		value, err := source.Get(ctx)
		if err != nil {
			return nil, err
		}
		*target = value
		return func(cmd *exec.Cmd) error {
			*target = ""
			return nil
		}, nil
	}
}

func EnvValue(name string, src *string) Customizer {
	return func(cmd *exec.Cmd) (Cleanup, error) {
		return FixedEnvValue(name, *src)(cmd)
	}
}

func FixedEnvValue(name string, value string) Customizer {
	return func(cmd *exec.Cmd) (Cleanup, error) {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", name, value))
		return nil, nil
	}
}
