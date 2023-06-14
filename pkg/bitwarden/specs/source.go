package specs

import (
	"github.com/masinger/incredible/pkg/specs/field"
)

type Source struct {
	Entry      string       `yaml:"entry" json:"entry"`
	Field      *field.Field `yaml:"field,omitempty" json:"field,omitempty"`
	Attachment *string      `yaml:"attachment,omitempty" json:"attachment,omitempty"`
}
