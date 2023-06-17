package source

import (
	"context"
	"io"
)

type Source interface {
}

type ValueSource interface {
	Source
	Get(ctx context.Context) (string, error)
}

type BinarySource interface {
	Source
	Read(ctx context.Context) (io.ReadCloser, error)
}

type StaticValueSource struct {
	Value *string
}

func (s StaticValueSource) Get(_ context.Context) (string, error) {
	return *s.Value, nil
}
