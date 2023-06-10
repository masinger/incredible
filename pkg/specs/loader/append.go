package loader

import (
	"github.com/masinger/incredible/pkg/specs"
	"path/filepath"
)

func AppendFileLoader(loader SpecsFileLoader, name string) SpecsFileLoader {
	return func(path string) (*specs.Manifest, error) {
		return loader(filepath.Join(path, name))
	}
}
