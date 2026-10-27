package processor

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"blc/internal/models"
)

func TestPublish(t *testing.T) {
	out := filepath.Join(t.TempDir(), "compiled.txt")
	p := NewBlocklistProcessor(Options{OutputFile: out, MinRules: 2})
	config := &models.BlocklistConfig{Name: "My list", Description: "test"}

	if err := p.publish(t.Context(), config, []byte("||a.com^\n||b.com^\n")); err != nil {
		t.Fatalf("publish: %v", err)
	}

	// Below the floor: rejected, and the previous list stays in place.
	err := p.publish(t.Context(), config, []byte("! header\n||a.com^\n\n"))
	if !errors.Is(err, ErrTooFewRules) {
		t.Fatalf("expected ErrTooFewRules, got %v", err)
	}

	got, err := os.ReadFile(filepath.Clean(out))
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"! Title: My list\n", "! Description: test\n", "! Rule count: 2\n", "! Last modified: ",
	} {
		if !strings.Contains(string(got), want) {
			t.Fatalf("header is missing %q:\n%s", want, got)
		}
	}

	if strings.Contains(string(got), "! Homepage:") {
		t.Fatalf("empty fields should be left out:\n%s", got)
	}

	if !strings.HasSuffix(string(got), "!\n||a.com^\n||b.com^\n") {
		t.Fatalf("previous list was replaced: %q", got)
	}

	// No temp files are left behind next to the output.
	entries, err := os.ReadDir(filepath.Dir(out))
	if err != nil {
		t.Fatal(err)
	}

	if len(entries) != 1 {
		t.Fatalf("expected only the output file, found %d entries", len(entries))
	}
}

func TestDropEmptySections(t *testing.T) {
	tests := []struct {
		name, input, expected string
	}{
		{"kept between rules", "a\n" + separator + "b\n", "a\n" + separator + "b\n"},
		{"trailing dropped", "a\n" + separator, "a\n"},
		{"leading dropped", separator + "b\n", "b\n"},
		{"consecutive collapse", "a\n" + separator + separator + "b\n", "a\n" + separator + "b\n"},
		{"without blank lines", "a\n! ----\nb\n! ----\n", "a\n! ----\nb\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := string(dropEmptySections([]byte(tt.input))); got != tt.expected {
				t.Fatalf("expected %q, got %q", tt.expected, got)
			}
		})
	}
}
