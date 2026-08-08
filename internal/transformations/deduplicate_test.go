package transformations

import (
	"bytes"
	"context"
	"testing"
)

func TestDeduplicate(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "remove exact duplicates",
			input: `example.com
malicious.com
example.com
tracking.com
malicious.com
`,
			expected: `example.com
malicious.com
tracking.com
`,
		},
		{
			name: "preserve order",
			input: `z.com
a.com
z.com
b.com
a.com
`,
			expected: `z.com
a.com
b.com
`,
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name: "no duplicates",
			input: `example.com
malicious.com
tracking.com
`,
			expected: `example.com
malicious.com
tracking.com
`,
		},
		{
			name: "case sensitive duplicates",
			input: `Example.com
example.com
EXAMPLE.COM
`,
			expected: `Example.com
example.com
EXAMPLE.COM
`,
		},
	}

	transform := &Deduplicate{}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := transform.Apply(context.TODO(), []byte(tt.input))
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if !bytes.Equal(output, []byte(tt.expected)) {
				t.Fatalf("Expected %q, got %q", tt.expected, output)
			}
		})
	}
}
