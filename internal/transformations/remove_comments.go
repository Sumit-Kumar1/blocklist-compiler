package transformations

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"strings"
)

// RemoveComments removes lines starting with ! (adblock) or # (hosts)
type RemoveComments struct{}

func (r RemoveComments) Apply(ctx context.Context, data []byte) ([]byte, error) {
	var buf strings.Builder

	if ctx.Err() != nil {
		return nil, errors.New("removeComments: context cancelled")
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, errors.New("removeComments: context cancelled")
		}

		line := scanner.Text()
		// Adblock comments start with '!', hosts with '#'
		if !strings.HasPrefix(line, "!") && !strings.HasPrefix(line, "#") {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return []byte(buf.String()), nil
}

func (r RemoveComments) Name() string {
	return "RemoveComments"
}
