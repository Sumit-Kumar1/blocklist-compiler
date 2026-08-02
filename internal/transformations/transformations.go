package transformations

import (
	"bufio"
	"bytes"
	"strings"
)

// RemoveComments removes lines starting with ! (adblock) or # (hosts)
type RemoveComments struct{}

func (r RemoveComments) Apply(data []byte) ([]byte, error) {
	var buf bytes.Builder
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
	return buf.Bytes(), nil
}

func (r RemoveComments) Name() string {
	return "RemoveComments"
}

// Deduplicate removes duplicate lines while preserving order
type Deduplicate struct{}

func (d Deduplicate) Apply(data []byte) ([]byte, error) {
	seen := make(map[string]bool)
	var buf bytes.Builder
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
	return buf.Bytes(), nil
}

func (d Deduplicate) Name() string {
	return "Deduplicate"
}

// Validate filters out invalid entries
type Validate struct{}

func (v Validate) Apply(data []byte) ([]byte, error) {
	var buf bytes.Builder
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
	return buf.Bytes(), nil
}

func (v Validate) Name() string {
	return "Validate"
}

// Compress trims whitespace from each line
type Compress struct{}

func (c Compress) Apply(data []byte) ([]byte, error) {
	var buf bytes.Builder
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
	return buf.Bytes(), nil
}

func (c Compress) Name() string {
	return "Compress"
}

// TrimLines trims whitespace from each line but preserves internal spacing
type TrimLines struct{}

func (t TrimLines) Apply(data []byte) ([]byte, error) {
	var buf bytes.Builder
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			buf.WriteString(trimmed)
			buf.WriteByte('\n')
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (t TrimLines) Name() string {
	return "TrimLines"
}

// RemoveEmptyLines removes empty lines
type RemoveEmptyLines struct{}

func (r RemoveEmptyLines) Apply(data []byte) ([]byte, error) {
	var buf bytes.Builder
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
	return buf.Bytes(), nil
}

func (r RemoveEmptyLines) Name() string {
	return "RemoveEmptyLines"
}

// InsertFinalNewLine ensures there's a final newline
type InsertFinalNewLine struct{}

func (i InsertFinalNewLine) Apply(data []byte) ([]byte, error) {
	if len(data) == 0 || data[len(data)-1] != '\n' {
		return append(data, '\n'), nil
	}
	return data, nil
}

func (i InsertFinalNewLine) Name() string {
	return "InsertFinalNewLine"
}

// InvertAllow converts allowlist entries to blocklist entries
type InvertAllow struct{}

func (i InvertAllow) Apply(data []byte) ([]byte, error) {
	var buf bytes.Builder
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
	return buf.Bytes(), nil
}

func (i InvertAllow) Name() string {
	return "InvertAllow"
}
