package transformations

import (
	"bufio"
	"bytes"
	"strings"
)

// InvertAllow converts allowlist entries to blocklist entries
type InvertAllow struct{}

func (i InvertAllow) Apply(data []byte) ([]byte, error) {
	var buf strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		// Convert @@||domain^ to ||domain^
		if strings.HasPrefix(line, "@@||") && strings.HasSuffix(line, "^") {
			line = "||" + strings.TrimPrefix(line, "@@||")
		}
		buf.WriteString(line)
		buf.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

func (i InvertAllow) Name() string {
	return "InvertAllow"
}
