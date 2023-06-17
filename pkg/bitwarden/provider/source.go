package provider

import (
	"context"
	"github.com/masinger/incredible/pkg/cli"
	"github.com/masinger/incredible/pkg/specs/field"
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

type valueSource struct {
	sessionKey    string
	itemId        string
	fieldAccessor field.Accessor[Item]
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
