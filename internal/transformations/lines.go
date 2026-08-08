package transformations

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
)

// maxLineSize bounds a single blocklist line. Public lists occasionally ship
// rules well past bufio's 64 KB default, which would otherwise abort the run.
const maxLineSize = 4 << 20

// mapLines feeds every line of data through fn and writes back the lines fn
// keeps. It is the shared body of every line-oriented transformation.
func mapLines(ctx context.Context, name string, data []byte, fn func(string) (string, bool)) ([]byte, error) {
	if ctx.Err() != nil {
		return nil, cancelled(name)
	}

	var buf strings.Builder

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxLineSize)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, cancelled(name)
		}

		line, keep := fn(scanner.Text())
		if !keep {
			continue
		}

		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", name, err)
	}

	return []byte(buf.String()), nil
}
