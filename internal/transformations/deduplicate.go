package transformations

import (
	"context"
)

// Deduplicate removes duplicate lines while preserving order
type Deduplicate struct{}

func (d Deduplicate) Apply(ctx context.Context, data []byte) ([]byte, error) {
	seen := make(map[string]struct{})

	return mapLines(ctx, d.Name(), data, func(line string) (string, bool) {
		if _, dup := seen[line]; dup {
			return "", false
		}

		seen[line] = struct{}{}

		return line, true
	})
}

func (d Deduplicate) Name() string {
	return "Deduplicate"
}
