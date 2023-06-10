package customizer

import "os/exec"

type Cleanup func(cmd *exec.Cmd) error
type Customizer func(cmd *exec.Cmd) (Cleanup, error)

func (c Customizer) Append(other Customizer) Customizer {
	return func(cmd *exec.Cmd) (Cleanup, error) {
		return Customizers{c, other}.Apply(cmd)
	}
}
