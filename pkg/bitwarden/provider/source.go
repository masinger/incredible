package provider

import (
	"context"
	"fmt"
	"github.com/masinger/incredible/pkg/cli"
	"io"
)

type fileSource struct {
	sessionKey     string
	itemId         string
	attachmentName string
}

func (f fileSource) Read(ctx context.Context) (io.ReadCloser, error) {
	cliClient := CliClient{
		cli.Cli{
			Ctx:        ctx,
			Executable: "bw",
		},
	}
	reader, err := cliClient.readAttachment(f.sessionKey, f.itemId, f.attachmentName)
	if reader == nil {
		return nil, err
	}
	return io.NopCloser(reader), err
}

type fieldAccessor func(item Item) (string, error)

var FieldAccessorPassword fieldAccessor = func(item Item) (string, error) {
	if item.Login == nil {
		return "", fmt.Errorf("no password found for entry")
	}
	return item.Login.Password, nil
}

var FieldAccessorUsername fieldAccessor = func(item Item) (string, error) {
	if item.Login == nil {
		return "", fmt.Errorf("no username found for entry")
	}
	return item.Login.Username, nil
}

type valueSource struct {
	sessionKey    string
	itemId        string
	fieldAccessor fieldAccessor
}

func (v valueSource) Get(ctx context.Context) (string, error) {
	cliClient := CliClient{
		cli.Cli{
			Ctx:        ctx,
			Executable: "bw",
		},
	}
	item, err := cliClient.getItem(v.sessionKey, v.itemId)
	if err != nil {
		return "", err
	}
	return v.fieldAccessor(*item)
}
