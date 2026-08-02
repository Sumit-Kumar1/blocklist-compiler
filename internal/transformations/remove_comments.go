package transformations

import (
	"bufio"
	"bytes"
	"strings"
)

// RemoveComments removes lines starting with ! (adblock) or # (hosts)
type RemoveComments struct{}

func (r RemoveComments) Apply(data []byte) ([]byte, error) {
	var buf strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
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
