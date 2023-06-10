package customizer

import (
	"context"
	"fmt"
	"github.com/masinger/incredible/pkg/provider/source"
	"os/exec"
)

func EnvValue(ctx context.Context, name string, source source.ValueSource) Customizer {
	return func(cmd *exec.Cmd) (Cleanup, error) {
		val, err := source.Get(ctx)
		if err != nil {
			return nil, err
		}
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", name, val))
		return nil, nil
	}
}
