package provider

import (
	"errors"
	"fmt"
)

var ErrProviderUnavailable = fmt.Errorf("the provider is unavailable")

func NewProviderUnavailableErr(reason error) error {
	return fmt.Errorf("the provider is unavailable: %w", errors.Join(reason, ErrProviderUnavailable))
}
