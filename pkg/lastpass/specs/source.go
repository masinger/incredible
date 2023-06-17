package specs

import "github.com/masinger/incredible/pkg/specs/field"

type LastPassAccountSource struct {
	AccountId string       `yaml:"id" json:"id"`
	Field     *field.Field `yaml:"field,omitempty" json:"field,omitempty"`
}
