package transformations

import (
	"bytes"
	"testing"
)

func TestTrimLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "trim leading and trailing whitespace",
			input: `  example.com
	malicious.com
   tracking.com
`,
			expected: `example.com
malicious.com
tracking.com
`,
		},
		{
			name: "preserve internal spacing",
			input: `127.0.0.1    localhost
0.0.0.0         tracking.com
`,
			expected: `127.0.0.1    localhost
0.0.0.0         tracking.com
`,
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:  "tabs and spaces",
			input: "\texample.com\t\n  malicious.com  \n",
			expected: `example.com
malicious.com
`,
		},
	}

	transform := &TrimLines{}
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
