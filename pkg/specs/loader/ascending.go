package loader

import (
	"github.com/masinger/incredible/pkg/specs"
	"path/filepath"
	"strings"
)

func AscendingFileLoader(loader SpecsFileLoader) SpecsFileLoader {
	var result SpecsFileLoader
	result = func(path string) (*specs.Manifest, error) {
		manifest, err := loader(path)
		if err != nil {
			return nil, err
		}
		if manifest != nil {
			return manifest, nil
		}
		parent := filepath.Dir(path)
		if strings.EqualFold(parent, path) {
			return nil, nil
		}
		return result(parent)
	}
	return result
}
