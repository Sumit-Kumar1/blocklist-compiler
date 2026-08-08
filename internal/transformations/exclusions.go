package transformations

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

// ExclusionFilter drops rules matching any of the configured patterns.
//
// Patterns mirror @adguard/hostlist-compiler and come in three forms, matched
// against the rule's full text:
//   - "/regex/"          — regular expression search
//   - "pattern*with*star" — wildcard, matching the whole rule
//   - "plainstring"       — substring containment
type ExclusionFilter struct {
	substrings []string
	patterns   []*regexp.Regexp
}

// NewExclusionFilter compiles patterns. Blank entries and "!" comments are ignored.
func NewExclusionFilter(patterns []string) (*ExclusionFilter, error) {
	filter := &ExclusionFilter{}

	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" || strings.HasPrefix(pattern, "!") {
			continue
		}

		compiled, err := compileExclusion(pattern)
		if err != nil {
			return nil, err
		}

		if compiled == nil {
			filter.substrings = append(filter.substrings, pattern)

			continue
		}

		filter.patterns = append(filter.patterns, compiled)
	}

	return filter, nil
}

// compileExclusion returns nil when the pattern is a plain substring.
func compileExclusion(pattern string) (*regexp.Regexp, error) {
	if len(pattern) > 1 && strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/") {
		compiled, err := regexp.Compile(strings.TrimSuffix(strings.TrimPrefix(pattern, "/"), "/"))
		if err != nil {
			return nil, fmt.Errorf("exclusion %q: %w", pattern, err)
		}

		return compiled, nil
	}

	if !strings.Contains(pattern, "*") {
		return nil, nil
	}

	// Wildcards match the whole rule, so anchor both ends.
	var expr strings.Builder

	expr.WriteString(`\A`)

	for i, literal := range strings.Split(pattern, "*") {
		if i > 0 {
			expr.WriteString(`.*`)
		}

		expr.WriteString(regexp.QuoteMeta(literal))
	}

	expr.WriteString(`\z`)

	compiled, err := regexp.Compile(expr.String())
	if err != nil {
		return nil, fmt.Errorf("exclusion %q: %w", pattern, err)
	}

	return compiled, nil
}

// Empty reports whether the filter would drop nothing.
func (f *ExclusionFilter) Empty() bool {
	return len(f.substrings) == 0 && len(f.patterns) == 0
}

// Excludes reports whether a rule matches any configured pattern.
func (f *ExclusionFilter) Excludes(rule string) bool {
	for _, substring := range f.substrings {
		if strings.Contains(rule, substring) {
			return true
		}
	}

	for _, pattern := range f.patterns {
		if pattern.MatchString(rule) {
			return true
		}
	}

	return false
}

// Apply drops every excluded rule, leaving comments and blank lines in place.
func (f *ExclusionFilter) Apply(ctx context.Context, data []byte) ([]byte, error) {
	if f.Empty() {
		return data, nil
	}

	return mapLines(ctx, f.Name(), data, func(line string) (string, bool) {
		if line == "" || isComment(line) {
			return line, true
		}

		return line, !f.Excludes(line)
	})
}

func (f *ExclusionFilter) Name() string {
	return "Exclusions"
}
