package transformations

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"strings"
)

// Deduplicate removes duplicate lines while preserving order
type Deduplicate struct{}

func (d Deduplicate) Apply(ctx context.Context, data []byte) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, errors.New("deduplicate: context cancelled")
	}

	seen := make(map[string]bool)

	var buf strings.Builder

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, errors.New("deduplicate: context cancelled")
		}

		line := scanner.Text()
		if !seen[line] {
			seen[line] = true
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return []byte(buf.String()), nil
}

func (d Deduplicate) Name() string {
	return "Deduplicate"
}
