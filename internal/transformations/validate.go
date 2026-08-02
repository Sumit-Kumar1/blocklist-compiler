package transformations

import (
	"bufio"
	"bytes"
	"strings"
)

// Validate filters out invalid entries
type Validate struct{}

func (v Validate) Apply(data []byte) ([]byte, error) {
	var buf strings.Builder
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		// Check for valid adblock format: ||domain^ or ||domain^$rules
		if strings.HasPrefix(line, "||") && strings.HasSuffix(line, "^") {
			buf.WriteString(line)
			buf.WriteByte('\n')
			continue
		}

		// Check for valid hosts format: IP domain
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			ip := parts[0]
			domain := parts[1]
			if ip != "" && domain != "" && strings.Contains(domain, ".") && !strings.HasPrefix(domain, "-") {
				buf.WriteString(line)
				buf.WriteByte('\n')
				continue
			}
		}

		// For hosts files, lines starting with # are comments
		if !strings.HasPrefix(line, "#") {
			buf.WriteString(line)
			buf.WriteByte('\n')
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return []byte(buf.String()), nil
}

func (v Validate) Name() string {
	return "Validate"
}
