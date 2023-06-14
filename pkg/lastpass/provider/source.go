package provider

import (
	"context"
	"fmt"
	"github.com/ansd/lastpass-go"
	"github.com/masinger/incredible/pkg/specs/field"
	"strings"
)

type accountSource struct {
	id            string
	provider      *Provider
	fieldAccessor field.Accessor[*lastpass.Account]
}

func (a accountSource) Get(ctx context.Context) (string, error) {
	accounts, err := a.provider.client.Accounts(ctx)
	if err != nil {
		return "", err
	}

	for _, account := range accounts {
		if strings.EqualFold(account.ID, a.id) {
			return a.fieldAccessor(account)
		}
	}

	return "", fmt.Errorf("account with id '%s' could not be found", a.id)
}
