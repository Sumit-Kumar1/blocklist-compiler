package transformations

import (
	"bytes"
	"context"
	"testing"
)

// Expectations verified against @adguard/hostlist-compiler v2.1.0.
func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "supported rules are kept",
			input: `||good.com^
||a.co^
||x7.com^$important
||x8.com^$dnsrewrite=1.2.3.4
||x9.com^$denyallow=a.com
||x10.com^$ctag=device_pc
||x11.com^$client=127.0.0.1
@@||allowed.com^
`,
			expected: `||good.com^
||a.co^
||x7.com^$important
||x8.com^$dnsrewrite=1.2.3.4
||x9.com^$denyallow=a.com
||x10.com^$ctag=device_pc
||x11.com^$client=127.0.0.1
@@||allowed.com^
`,
		},
		{
			name: "public-suffix wildcards are dropped, ordinary wildcards kept",
			input: `||*.org^
||*.com^
||*.co.uk^
||*.example.com^
||*.sub.example.com^
||keep.com^
`,
			expected: `||*.example.com^
||*.sub.example.com^
||keep.com^
`,
		},
		{
			name:     "dotless hosts are dropped",
			input:    "||localhost^\n||ab.cd^\n",
			expected: "||ab.cd^\n",
		},
		{
			name: "IP rules and dotless hosts are dropped",
			input: `1.2.3.4
||1.2.3.4^
::1
||ab^
||keep.com^
`,
			expected: "||keep.com^\n",
		},
		{
			name: "cosmetic rules are dropped",
			input: `example.org##.banner
example.org#@#.banner
||keep.com^
`,
			expected: "||keep.com^\n",
		},
		{
			name: "rules with unsupported modifiers are dropped",
			input: `||x.com^$third-party
||x2.com^$3p
||x3.com^$document
||x4.com^$popup
||x5.com^$all
||x6.com^$network
||keep.com^
`,
			expected: "||keep.com^\n",
		},
		{
			name: "hosts entries and comments pass through",
			input: `0.0.0.0 blocked-hosts.com
# hosts comment
plain-domain.com
`,
			expected: `0.0.0.0 blocked-hosts.com
# hosts comment
plain-domain.com
`,
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
	}

	transform := &Validate{}

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
