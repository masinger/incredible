package loader

import (
	"errors"
	"github.com/masinger/incredible/pkg/specs"
	"gopkg.in/yaml.v3"
	"os"
)

func NewYamlLoader() SpecsFileLoader {
	return func(path string) (res *specs.Manifest, err error) {
		f, err := os.Open(path)
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return nil, nil
			}
			return nil, err
		}
		defer func() {
			if closeErr := f.Close(); closeErr != nil {
				err = errors.Join(err, closeErr)
			}
		}()
		dec := yaml.NewDecoder(f)
		var manifest specs.Manifest
		err = dec.Decode(&manifest)
		if err != nil {
			return nil, err
		}
		return &manifest, nil
	}
}
