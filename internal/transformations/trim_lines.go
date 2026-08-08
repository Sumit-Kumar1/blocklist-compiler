package transformations

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"strings"
)

// TrimLines trims whitespace from each line but preserves internal spacing
type TrimLines struct{}

func (t TrimLines) Apply(ctx context.Context, data []byte) ([]byte, error) {
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
		// Trim leading/trailing whitespace but preserve internal spacing
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			buf.WriteString(trimmed)
			buf.WriteByte('\n')
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return []byte(buf.String()), nil
}

func (t TrimLines) Name() string {
	return "TrimLines"
}
