package transformations

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"strings"
)

// Compress trims whitespace from each line
type Compress struct{}

func (c Compress) Apply(ctx context.Context, data []byte) ([]byte, error) {
	var buf strings.Builder

	if ctx.Err() != nil {
		return nil, errors.New("compress: context cancelled")
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, errors.New("compress: context cancelled")
		}

		line := strings.TrimSpace(scanner.Text())
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

func (c Compress) Name() string {
	return "Compress"
}
