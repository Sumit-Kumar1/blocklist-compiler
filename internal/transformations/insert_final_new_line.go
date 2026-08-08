package transformations

import (
	"context"
	"errors"
)

// InsertFinalNewLine ensures there's a final newline
type InsertFinalNewLine struct{}

func (i InsertFinalNewLine) Apply(ctx context.Context, data []byte) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, errors.New("insertFinalNewLine: context cancelled")
	}

	if len(data) == 0 || data[len(data)-1] != '\n' {
		return append(data, '\n'), nil
	}

	return data, nil
}

func (i InsertFinalNewLine) Name() string {
	return "InsertFinalNewLine"
}
