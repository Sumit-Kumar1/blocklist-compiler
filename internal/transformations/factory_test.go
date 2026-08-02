package transformations

import (
	"testing"
)

func TestTransformationFactory(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{"RemoveComments", "RemoveComments", false},
		{"Deduplicate", "Deduplicate", false},
		{"Validate", "Validate", false},
		{"Compress", "Compress", false},
		{"TrimLines", "TrimLines", false},
		{"RemoveEmptyLines", "RemoveEmptyLines", false},
		{"InsertFinalNewLine", "InsertFinalNewLine", false},
		{"InvertAllow", "InvertAllow", false},
		{"Unknown", "Unknown", true},
	}

	factory := &TransformationFactory{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transform, err := factory.Create(tt.input)
			if tt.expectError {
				if err == nil {
					t.Fatal("Expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if transform.Name() != tt.input {
				t.Fatalf("Expected name %q, got %q", tt.input, transform.Name())
			}
		})
	}
}
