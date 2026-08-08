package transformations

import (
	"bytes"
	"context"
	"testing"
)

const exclusionInput = `||tailscale.com^
||sub.tailscale.com^
||nottailscale.com^
||luciferdonghua.in^
||keepme.com^
||ads.example.net^
`

// Expectations verified against @adguard/hostlist-compiler v2.1.0.
func TestExclusionFilter(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		expected string
	}{
		{
			name:     "plain pattern matches as a substring, not a domain suffix",
			patterns: []string{"||tailscale.com^"},
			expected: `||sub.tailscale.com^
||nottailscale.com^
||luciferdonghua.in^
||keepme.com^
||ads.example.net^
`,
		},
		{
			name:     "plain pattern without a caret still matches",
			patterns: []string{"||luciferdonghua.in"},
			expected: `||tailscale.com^
||sub.tailscale.com^
||nottailscale.com^
||keepme.com^
||ads.example.net^
`,
		},
		{
			name:     "bare substring matches everywhere it appears",
			patterns: []string{"tailscale"},
			expected: `||luciferdonghua.in^
||keepme.com^
||ads.example.net^
`,
		},
		{
			name:     "wildcards match the whole rule",
			patterns: []string{"*ads*"},
			expected: `||tailscale.com^
||sub.tailscale.com^
||nottailscale.com^
||luciferdonghua.in^
||keepme.com^
`,
		},
		{
			name:     "an unanchored wildcard matches nothing",
			patterns: []string{"ads*"},
			expected: exclusionInput,
		},
		{
			name:     `regex patterns are matched as a search`,
			patterns: []string{`/^\|\|sub\./`},
			expected: `||tailscale.com^
||nottailscale.com^
||luciferdonghua.in^
||keepme.com^
||ads.example.net^
`,
		},
		{
			name:     "no patterns keeps everything",
			patterns: nil,
			expected: exclusionInput,
		},
		{
			name:     "blank entries and comments are ignored",
			patterns: []string{"", "  ", "! a comment"},
			expected: exclusionInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := NewExclusionFilter(tt.patterns)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			output, err := filter.Apply(context.Background(), []byte(exclusionInput))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !bytes.Equal(output, []byte(tt.expected)) {
				t.Fatalf("Expected %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestExclusionFilterRejectsBadRegex(t *testing.T) {
	if _, err := NewExclusionFilter([]string{"/([unclosed/"}); err == nil {
		t.Fatal("Expected an error for an invalid regex, got nil")
	}
}
