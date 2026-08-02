package transformations

import (
	"bufio"
	"bytes"
	"strings"
)

// Compress trims whitespace from each line
type Compress struct{}

func (c Compress) Apply(data []byte) ([]byte, error) {
	var buf strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
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
