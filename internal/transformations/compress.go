package transformations

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"strings"
)

// Compress converts hosts-format and bare-domain lines into adblock rules, then
// discards rules made redundant by a broader rule elsewhere in the list.
//
// It mirrors @adguard/hostlist-compiler's Compress transformation.
type Compress struct{}

func (c Compress) Apply(ctx context.Context, data []byte) ([]byte, error) {
	lines, err := c.normalize(ctx, data)
	if err != nil {
		return nil, err
	}

	// A plain "||domain^" rule subsumes every rule for its subdomains.
	parents := make(map[string]struct{}, len(lines))

	for _, line := range lines {
		if domain, ok := blockedDomain(line); ok {
			parents[domain] = struct{}{}
		}
	}

	var (
		buf  strings.Builder
		seen = make(map[string]struct{}, len(lines))
	)

	for _, line := range lines {
		if ctx.Err() != nil {
			return nil, cancelled(c.Name())
		}

		if !isComment(line) {
			if _, dup := seen[line]; dup {
				continue
			}

			seen[line] = struct{}{}

			if domain, ok := blockedDomain(line); ok && hasBroaderParent(domain, parents) {
				continue
			}
		}

		buf.WriteString(line)
		buf.WriteByte('\n')
	}

	return []byte(buf.String()), nil
}

// normalize rewrites hosts entries and bare domains as "||domain^" rules and
// leaves every other line untouched.
func (c Compress) normalize(ctx context.Context, data []byte) ([]string, error) {
	if ctx.Err() != nil {
		return nil, cancelled(c.Name())
	}

	var lines []string

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 0, bufio.MaxScanTokenSize), maxLineSize)

	for scanner.Scan() {
		if ctx.Err() != nil {
			return nil, cancelled(c.Name())
		}

		line := scanner.Text()

		switch {
		case isComment(line):
			lines = append(lines, line)
		case isPlainDomain(line):
			lines = append(lines, "||"+line+"^")
		default:
			domains, ok := hostsDomains(line)
			if !ok {
				lines = append(lines, line)

				continue
			}

			for _, domain := range domains {
				lines = append(lines, "||"+domain+"^")
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("%s: %w", c.Name(), err)
	}

	return lines, nil
}

// hasBroaderParent reports whether an ancestor domain is itself blocked outright.
func hasBroaderParent(domain string, parents map[string]struct{}) bool {
	for rest := domain; ; {
		_, cut, found := strings.Cut(rest, ".")
		if !found {
			return false
		}

		rest = cut
		if _, ok := parents[rest]; ok && coversSubdomain(rest, domain) {
			return true
		}
	}
}

func (c Compress) Name() string {
	return "Compress"
}
