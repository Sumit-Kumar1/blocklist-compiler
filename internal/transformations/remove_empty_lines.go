package transformations

import (
	"bufio"
	"bytes"
	"strings"
)

// RemoveEmptyLines removes empty lines
type RemoveEmptyLines struct{}

func (r RemoveEmptyLines) Apply(data []byte) ([]byte, error) {
	var buf strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
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
