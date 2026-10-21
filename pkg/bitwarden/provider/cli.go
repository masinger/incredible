package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/masinger/incredible/pkg/cli"
	"io"
	"os"
	"time"
)

type CliClient struct {
	cli.Cli
}

func newCliClient(ctx context.Context, timeout time.Duration) CliClient {
	return CliClient{
		cli.Cli{
			Ctx:        ctx,
			Executable: "bw",
			Timeout:    timeout,
		},
	}
}

func (b CliClient) isLoggedIn() (bool, error) {
	return cli.Succeeded(b.Cli.RunInheritStreams("login", "--check"))
}

func (b CliClient) login() (string, error) {
	// Prompts for the master password and 2FA, so no timeout.
	return b.Cli.WithoutTimeout().RunReadStdout("login", "--raw")
}

func (b CliClient) isUnlocked() (bool, error) {
	return cli.Succeeded(b.Cli.RunInheritStreams("unlock", "--check"))
}

func (b CliClient) unlock() (string, error) {
	// Prompts for the master password, so no timeout.
	return b.Cli.WithoutTimeout().RunReadStdout("unlock", "--raw")
}

func (b CliClient) activeSession() (string, error) {
	loggedIn, err := b.isLoggedIn()
	if err != nil {
		return "", err
	}
	if !loggedIn {
		return b.login()
	}
	unlocked, err := b.isUnlocked()
	if err != nil {
		return "", err
	}
	if !unlocked {
		return b.unlock()
	}
	sessionKey := os.Getenv("BW_SESSION")
	if sessionKey == "" {
		return "", fmt.Errorf("could not obtain session key")
	}
	return sessionKey, nil
}

func (b CliClient) getItem(sessionKey string, itemId string) (*Item, error) {
	jsonString, err := b.Cli.WithEnv("BW_SESSION", sessionKey).RunReadStdout(
		"get",
		"item",
		itemId,
		"--raw",
	)
	if err != nil {
		return nil, err
	}
	var item Item
	err = json.Unmarshal([]byte(jsonString), &item)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (b CliClient) readAttachment(
	sessionKey string,
	itemId string,
	attachmentName string,
) (io.Reader, error) {
	return b.Cli.WithEnv("BW_SESSION", sessionKey).RunRedirectStdout(
		"get",
		"attachment",
		attachmentName,
		"--itemid",
		itemId,
		"--raw",
	)
}
