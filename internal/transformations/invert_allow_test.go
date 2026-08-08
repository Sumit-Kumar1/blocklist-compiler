package transformations

import (
	"bytes"
	"context"
	"testing"
)

func TestInvertAllow(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "invert adblock allow to block",
			input: `@@||example.com^
@@||malicious.com^$important
`,
			expected: `||example.com^
||malicious.com^$important
`,
		},
		{
			name: "preserve block entries",
			input: `||example.com^
||malicious.com^$important
`,
			expected: `||example.com^
||malicious.com^$important
`,
		},
		{
			name: "mixed allow and block",
			input: `@@||whitelist.com^
||block.com^
@@||allow.com^$script
`,
			expected: `||whitelist.com^
||block.com^
||allow.com^$script
`,
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name: "non-standard allow format",
			input: `@@example.com
@@malicious.com^$third-party
`,
			expected: `@@example.com
@@malicious.com^$third-party
`,
		},
	}

	transform := &InvertAllow{}
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
