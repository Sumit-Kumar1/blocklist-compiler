package transformations

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"strings"
)

// RemoveEmptyLines removes empty lines
type RemoveEmptyLines struct{}

func (r RemoveEmptyLines) Apply(ctx context.Context, data []byte) ([]byte, error) {
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
		if line != "" {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return []byte(buf.String()), nil
}

func (r RemoveEmptyLines) Name() string {
	return "RemoveEmptyLines"
}
