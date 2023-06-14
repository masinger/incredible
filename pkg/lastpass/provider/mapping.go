package provider

import (
	"github.com/ansd/lastpass-go"
	"github.com/masinger/incredible/pkg/specs/field"
)

var fieldMapping = &field.Mapping[*lastpass.Account]{
	Username: func(element *lastpass.Account) (string, error) {
		return element.Username, nil
	},
	Password: func(element *lastpass.Account) (string, error) {
		return element.Password, nil
	},
}
