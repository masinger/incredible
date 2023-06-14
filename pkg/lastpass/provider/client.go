package provider

import (
	"context"
	"github.com/ansd/lastpass-go"
	"github.com/masinger/incredible/internal/interactive"
	"github.com/masinger/incredible/pkg/logging"
)

type ClientCreator func(ctx context.Context) (*lastpass.Client, error)

var UsernameInput = interactive.Input{
	Message:        "Username",
	DisableDefault: true,
}

var PasswordInput = interactive.Input{
	Message:        "Password",
	Masked:         true,
	DisableDefault: true,
}

var InteractiveClientCreator ClientCreator = func(ctx context.Context) (*lastpass.Client, error) {
	username, err := logging.Interactive.Input(UsernameInput)
	if err != nil {
		return nil, err
	}
	password, err := logging.Interactive.Input(PasswordInput)
	if err != nil {
		return nil, err
	}

	return lastpass.NewClient(ctx, username, password)
}
