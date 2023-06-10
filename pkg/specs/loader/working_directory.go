package loader

import (
	"github.com/masinger/incredible/pkg/specs"
	"os"
)

func CurrentWorkingDirectoryLoader(loader SpecsFileLoader) SpecsLoader {
	return func() (*specs.Manifest, error) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, err
		}
		return loader(wd)
	}
}
