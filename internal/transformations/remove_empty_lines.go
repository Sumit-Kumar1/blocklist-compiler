package transformations

import (
	"context"
)

// RemoveEmptyLines removes empty lines
type RemoveEmptyLines struct{}

func (r RemoveEmptyLines) Apply(ctx context.Context, data []byte) ([]byte, error) {
	return mapLines(ctx, r.Name(), data, func(line string) (string, bool) {
		return line, line != ""
	})
}

func (r RemoveEmptyLines) Name() string {
	return "RemoveEmptyLines"
}
