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
	loggedIn, err := cli.Succeeded(err)
	if err != nil || loggedIn {
		return err
	}
	// Opens a browser or device code flow, so no timeout.
	return c.Cli.WithoutTimeout().RunInheritStreams("login")
}

func (c CliClient) readSecretById(id string) (*Secret, error) {
	// Waits for az to exit, so it isn't killed while still running, and its
	// exit status is checked.
	output, err := c.Cli.RunReadStdout(
		"keyvault",
		"secret",
		"show",
		"--id",
		id,
	)
	if err != nil {
		return nil, err
	}
	var secret Secret
	if err = json.Unmarshal([]byte(output), &secret); err != nil {
		return nil, err
	}
	return &secret, nil
}
