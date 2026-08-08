package transformations

import (
	"bytes"
	"context"
	"testing"
)

// Expectations verified against @adguard/hostlist-compiler v2.1.0.
func TestInvertAllow(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "blocking rules become allow rules",
			input: `||example.com^
||malicious.com^$important
`,
			expected: `@@||example.com^
@@||malicious.com^$important
`,
		},
		{
			name:     "bare domains are prefixed too",
			input:    "plainword\naccount.xiaomi.com\n",
			expected: "@@plainword\n@@account.xiaomi.com\n",
		},
		{
			name:     "existing allow rules are left alone",
			input:    "@@||allowed.com^\n",
			expected: "@@||allowed.com^\n",
		},
		{
			name: "hosts-format entries are left alone",
			input: `0.0.0.0 hosts-style.com
1.2.3.4 another.com
::1 ipv6host.com
`,
			expected: `0.0.0.0 hosts-style.com
1.2.3.4 another.com
::1 ipv6host.com
`,
		},
		{
			name:     "comments are left alone",
			input:    "! adblock comment\n# hosts comment\n",
			expected: "! adblock comment\n# hosts comment\n",
		},
		{
			name:     "empty input",
			input:    "",
			expected: "",
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

// TestInvertAllowXiaomiChain covers the RemoveComments -> Compress -> InvertAllow
// pipeline the Xiaomi allowlists use, which must yield allow rules.
func TestInvertAllowXiaomiChain(t *testing.T) {
	pipeline := &TransformationPipeline{}
	pipeline.Add(&RemoveComments{})
	pipeline.Add(&Compress{})
	pipeline.Add(&InvertAllow{})

	input := "! comment\naccount.xiaomi.com\napi.io.mi.com\n"
	expected := "@@||account.xiaomi.com^\n@@||api.io.mi.com^\n"

	output, err := pipeline.Apply(context.Background(), []byte(input))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !bytes.Equal(output, []byte(expected)) {
		t.Fatalf("Expected %q, got %q", expected, output)
	}
}
