package transformations

import (
	"bufio"
	"bytes"
	"strings"
)

// Deduplicate removes duplicate lines while preserving order
type Deduplicate struct{}

func (d Deduplicate) Apply(data []byte) ([]byte, error) {
	seen := make(map[string]bool)
	var buf strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
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
