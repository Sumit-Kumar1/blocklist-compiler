package transformations

import (
	"bufio"
	"bytes"
	"strings"
)

type RemoveComments struct{}

func (r RemoveComments) Name() string {
	return "RemoveComments"
}

func (r RemoveComments) Apply(data []byte) ([]byte, error) {
	var sb strings.Builder

	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()

		if !strings.HasPrefix(line, "!") && !strings.HasPrefix(line, "#") {
			sb.WriteString(line)
			sb.WriteByte('\n')
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return []byte(sb.String()), nil
}
