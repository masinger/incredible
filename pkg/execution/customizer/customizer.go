package customizer

import "os/exec"

type Cleanup func(cmd *exec.Cmd) error
type Customizer func(cmd *exec.Cmd) (Cleanup, error)
