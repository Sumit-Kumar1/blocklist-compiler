package transformations

import (
	"bytes"
	"context"
	"testing"
)

// Expectations verified against @adguard/hostlist-compiler v2.1.0.
func TestCompress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "hosts entries become adblock rules",
			input: `0.0.0.0 blocked-hosts.com
127.0.0.1 another-hosts.com
`,
			expected: `||blocked-hosts.com^
||another-hosts.com^
`,
		},
		{
			name:     "one hosts line with several domains splits",
			input:    "0.0.0.0 multi-a.com multi-b.com\n",
			expected: "||multi-a.com^\n||multi-b.com^\n",
		},
		{
			name:     "bare domain becomes an adblock rule",
			input:    "plain-domain.com\n",
			expected: "||plain-domain.com^\n",
		},
		{
			name: "existing adblock rules are untouched",
			input: `||adblock-rule.com^
||with-modifier.com^$important
@@||existing-allow.com^
||*.org^
example.org##.banner
`,
			expected: `||adblock-rule.com^
||with-modifier.com^$important
@@||existing-allow.com^
||*.org^
example.org##.banner
`,
		},
		{
			name: "subdomains covered by a plain parent are dropped",
			input: `||parent.com^
||sub.parent.com^
||deep.sub.parent.com^
||notparent.com^
`,
			expected: `||parent.com^
||notparent.com^
`,
		},
		{
			name: "a parent carrying modifiers does not cover its subdomains",
			input: `||mod.com^$important
||sub.mod.com^
`,
			expected: `||mod.com^$important
||sub.mod.com^
`,
		},
		{
			name:     "duplicates collapse",
			input:    "||other.com^\n||other.com^\n",
			expected: "||other.com^\n",
		},
		{
			name:     "comments survive",
			input:    "# hosts comment\n! adblock comment\n",
			expected: "# hosts comment\n! adblock comment\n",
		},
		{
			name:     "indented lines are not parsed as domains",
			input:    "  leading-space.com\n",
			expected: "  leading-space.com\n",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
	}

	transform := &Compress{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := transform.Apply(context.Background(), []byte(tt.input))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !bytes.Equal(output, []byte(tt.expected)) {
				t.Fatalf("Expected %q, got %q", tt.expected, output)
			}
		})
	}
}
