package provider

import (
	"context"
)

type secretSource struct {
	itemId   string
	provider *Provider
}

func (s secretSource) Get(ctx context.Context) (string, error) {
	cli := s.provider.client(ctx)
	secret, err := cli.readSecretById(s.itemId)
	if err != nil {
		return "", err
	}
	return secret.Value, nil
}
