package specs

import "github.com/masinger/incredible/pkg/specs/mapping"

type Asset struct {
	Src      Source            `yaml:"src" json:"src"`
	Mappings []mapping.Mapping `yaml:"mappings" json:"mappings"`
}
