package provider

import (
	"context"
	"github.com/masinger/incredible/pkg/specs/field"
	"io"
	"time"
)

type fileSource struct {
	sessionKey     string
	itemId         string
	attachmentName string
	timeout        time.Duration
}

func (f fileSource) Read(ctx context.Context) (io.ReadCloser, error) {
	cliClient := newCliClient(ctx, f.timeout)
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
	timeout       time.Duration
}

func (v valueSource) Get(ctx context.Context) (string, error) {
	cliClient := newCliClient(ctx, v.timeout)
	item, err := cliClient.getItem(v.sessionKey, v.itemId)
	if err != nil {
		return "", err
	}
	return v.fieldAccessor(*item)
}
