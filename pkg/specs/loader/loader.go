package loader

import (
	"github.com/masinger/incredible/pkg/specs"
)

type SpecsLoader func() (*specs.Manifest, error)
type SpecsFileLoader func(path string) (*specs.Manifest, error)
