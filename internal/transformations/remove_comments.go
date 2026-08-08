package transformations

import (
	"context"
	"strings"
)

// RemoveComments removes lines starting with ! (adblock) or # (hosts)
type RemoveComments struct{}

func (r RemoveComments) Apply(ctx context.Context, data []byte) ([]byte, error) {
	return mapLines(ctx, r.Name(), data, func(line string) (string, bool) {
		// Adblock comments start with '!', hosts with '#', and a list header
		// such as "[Adblock Plus 2.0]" is metadata rather than a rule.
		return line, !isComment(line) && !strings.HasPrefix(line, "[")
	})
}

func (r RemoveComments) Name() string {
	return "RemoveComments"
}
