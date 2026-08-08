package transformations

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"strings"
)

// InvertAllow converts allowlist entries to blocklist entries
type InvertAllow struct{}

func (i InvertAllow) Apply(ctx context.Context, data []byte) ([]byte, error) {
	var buf strings.Builder

	if ctx.Err() != nil {
		return nil, errors.New("invertAllow: context cancelled")
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, errors.New("invertAllow: context cancelled")
		}

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
