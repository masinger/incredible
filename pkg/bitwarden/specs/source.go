package specs

type Field string

var FieldPassword Field = "password"
var FieldUsername Field = "username"

type Source struct {
	Entry      string  `yaml:"entry" json:"entry"`
	Field      *Field  `yaml:"field,omitempty" json:"field,omitempty"`
	Attachment *string `yaml:"attachment,omitempty" json:"attachment,omitempty"`
}
