package transformations

import (
	"bytes"
	"context"
	"testing"
)

func TestInsertFinalNewLine(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "add newline to data without final newline",
			input:    "example.com\nmalicious.com",
			expected: "example.com\nmalicious.com\n",
		},
		{
			name:     "preserve existing final newline",
			input:    "example.com\nmalicious.com\n",
			expected: "example.com\nmalicious.com\n",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "\n",
		},
		{
			name:     "single line without newline",
			input:    "example.com",
			expected: "example.com\n",
		},
		{
			name:     "single line with newline",
			input:    "example.com\n",
			expected: "example.com\n",
		},
	}

	transform := &InsertFinalNewLine{}

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
