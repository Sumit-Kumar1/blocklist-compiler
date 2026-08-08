package transformations

import (
	"context"
	"strings"
)

// TrimLines removes leading and trailing whitespace, mirroring
// @adguard/hostlist-compiler. Emptied lines are left for RemoveEmptyLines.
type TrimLines struct{}

func (t TrimLines) Apply(ctx context.Context, data []byte) ([]byte, error) {
	return mapLines(ctx, t.Name(), data, func(line string) (string, bool) {
		return strings.TrimSpace(line), true
	})
}

func (t TrimLines) Name() string {
	return "TrimLines"
}
