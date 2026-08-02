package transformations

import (
	"bytes"
	"testing"
)

func TestRemoveComments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "remove adblock comments",
			input: `! This is a comment
example.com
! Another comment
malicious.com
`,
			expected: `example.com
malicious.com
`,
		},
		{
			name: "remove hosts comments",
			input: `# This is a comment
127.0.0.1 localhost
# Another comment
0.0.0.0 tracking.com
`,
			expected: `127.0.0.1 localhost
0.0.0.0 tracking.com
`,
		},
		{
			name: "mixed comments",
			input: `! Adblock comment
example.com
# Hosts comment
malicious.com
`,
			expected: `example.com
malicious.com
`,
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name:     "only comments",
			input:    "! Comment 1\n# Comment 2\n",
			expected: "",
		},
	}

	transform := &RemoveComments{}
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
