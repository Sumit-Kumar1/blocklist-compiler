package transformations

import (
	"bytes"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "valid adblock entries",
			input: `||example.com^
||malicious.com^$important
@@||whitelist.com^
`,
			expected: `||example.com^
||malicious.com^$important
@@||whitelist.com^
`,
		},
		{
			name: "valid hosts entries",
			input: `127.0.0.1 localhost
0.0.0.0 tracking.com
`,
			expected: `127.0.0.1 localhost
0.0.0.0 tracking.com
`,
		},
		{
			name: "invalid entries filtered",
			input: `||example.com^
invalid-entry
||malicious.com^
not-a-domain
`,
			expected: `||example.com^
||malicious.com^
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
			output, err := transform.Apply([]byte(tt.input))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !bytes.Equal(output, []byte(tt.expected)) {
				t.Fatalf("Expected %q, got %q", tt.expected, output)
			}
		})
	}
}
