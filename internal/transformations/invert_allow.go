package transformations

import (
	"context"
	"strings"
)

// InvertAllow converts blocking rules into allow rules, mirroring
// @adguard/hostlist-compiler. Comments, blank lines, hosts-format entries and
// rules that already allow are passed through untouched.
type InvertAllow struct{}

func (i InvertAllow) Apply(ctx context.Context, data []byte) ([]byte, error) {
	return mapLines(ctx, i.Name(), data, func(line string) (string, bool) {
		if line == "" || isComment(line) || strings.HasPrefix(line, "@@") {
			return line, true
		}

		if _, isHosts := hostsDomains(line); isHosts {
			return line, true
		}

		return "@@" + line, true
	})
}

func (i InvertAllow) Name() string {
	return "InvertAllow"
}
