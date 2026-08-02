package transformations

import (
	"bytes"
	"testing"
)

func TestTransformationPipeline(t *testing.T) {
	tests := []struct {
		name        string
		transforms  []Transformation
		input       string
		expected    string
		expectError bool
	}{
		{
			name: "single transformation",
			transforms: []Transformation{
				&RemoveComments{},
			},
			input: `! Comment
example.com
`,
			expected: `example.com
`,
		},
		{
			name: "multiple transformations",
			transforms: []Transformation{
				&RemoveComments{},
				&Deduplicate{},
				&RemoveEmptyLines{},
			},
			input: `! Comment
example.com

example.com
`,
			expected: `example.com
`,
		},
		{
			name: "order matters",
			transforms: []Transformation{
				&RemoveComments{},
				&Deduplicate{},
			},
			input: `! Comment
example.com
example.com
`,
			expected: `example.com
`,
		},
		{
			name:       "empty pipeline",
			transforms: []Transformation{},
			input: `example.com
`,
			expected: `example.com
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pipeline := &TransformationPipeline{}
			for _, t := range tt.transforms {
				pipeline.Add(t)
			}

			output, err := pipeline.Apply([]byte(tt.input))
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if !bytes.Equal(output, []byte(tt.expected)) {
				t.Fatalf("Expected %q, got %q", tt.expected, output)
			}
		})
	}
}
