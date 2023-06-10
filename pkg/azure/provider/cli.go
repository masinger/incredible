package provider

import (
	"encoding/json"
	"github.com/masinger/incredible/pkg/cli"
)

type CliClient struct {
	cli.Cli
}

func (c CliClient) login() error {
	// Check if we are logged in
	_, err := c.Cli.RunReadStdout("account", "list")
	if err == nil {
		return nil
	}
	return c.Cli.RunInheritStreams("login")
}

func (c CliClient) readSecretById(id string) (*Secret, error) {
	reader, err := c.Cli.RunRedirectStdout(
		"keyvault",
		"secret",
		"show",
		"--id",
		id,
	)
	if err != nil {
		return nil, err
	}
	dec := json.NewDecoder(reader)
	var secret Secret
	if err = dec.Decode(&secret); err != nil {
		return nil, err
	}
	return &secret, err
}
