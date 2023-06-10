package loader

import (
	"github.com/masinger/incredible/pkg/specs"
)

func PriorityFileLoader(loaders ...SpecsFileLoader) SpecsFileLoader {
	return func(path string) (*specs.Manifest, error) {
		for _, l := range loaders {
			manifest, err := l(path)
			if err != nil {
				return nil, err
			}
			if manifest != nil {
				return manifest, nil
			}
		}
		return nil, nil
	}
}

func PriorityLoader(loaders ...SpecsLoader) SpecsLoader {
	return func() (*specs.Manifest, error) {
		for _, l := range loaders {
			manifest, err := l()
			if err != nil {
				return nil, err
			}
			if manifest != nil {
				return manifest, nil
			}
		}
		return nil, nil
	}
}
