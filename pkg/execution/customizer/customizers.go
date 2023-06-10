package customizer

import (
	"errors"
	"os/exec"
)

type Customizers []Customizer

func (c Customizers) Apply(cmd *exec.Cmd) (Cleanup, error) {
	var cleanups []Cleanup
	var cleanup Cleanup = func(cmd *exec.Cmd) error {
		var cleanupErrors []error
		for _, currentCleanup := range cleanups {
			cleanupErr := currentCleanup(cmd)
			if cleanupErr != nil {
				cleanupErrors = append(cleanupErrors, cleanupErr)
			}
		}
		return errors.Join(cleanupErrors...)
	}
	for _, customizer := range c {
		cl, err := customizer(cmd)
		if cl != nil {
			cleanups = append(cleanups, cl)
		}
		if err != nil {
			return cleanup, err
		}
	}
	return cleanup, nil
}
