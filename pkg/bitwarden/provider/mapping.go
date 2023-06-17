package provider

import (
	"fmt"
	"github.com/masinger/incredible/pkg/specs/field"
)

var fieldMapping = &field.Mapping[Item]{
	Username: func(item Item) (string, error) {
		if item.Login == nil {
			return "", fmt.Errorf("no password found for entry")
		}
		return item.Login.Username, nil
	},
	Password: func(item Item) (string, error) {
		if item.Login == nil {
			return "", fmt.Errorf("no username found for entry")
		}
		return item.Login.Password, nil
	},
}
