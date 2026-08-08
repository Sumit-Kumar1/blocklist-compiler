package transformations

import (
	"bytes"
	"context"
	"testing"
)

func TestRemoveEmptyLines(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "remove empty lines",
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
			name: "remove lines with only whitespace",
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
			name:     "empty input",
			input:    "",
			expected: "",
		},
		{
			name: "no empty lines",
			input: `example.com
malicious.com
tracking.com
`,
			expected: `example.com
malicious.com
tracking.com
`,
		},
	}

	transform := &RemoveEmptyLines{}

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
