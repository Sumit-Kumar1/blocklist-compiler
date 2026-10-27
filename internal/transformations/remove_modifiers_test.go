package transformations

import (
	"bytes"
	"context"
	"testing"
)

func TestRemoveModifiers(t *testing.T) {
	input := `! comment $third-party
||a.com^$third-party
||b.com^$3p,important
||c.com^$important,popup,doc
||d.com^$~third-party
||e.com^
@@||f.com^$document
/ads$/$network
/ads$/
0.0.0.0 g.com
`
	expected := `! comment $third-party
||a.com^
||b.com^$important
||c.com^$important
||d.com^$~third-party
||e.com^
@@||f.com^
/ads$/
/ads$/
0.0.0.0 g.com
`

	output, err := RemoveModifiers{}.Apply(context.Background(), []byte(input))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !bytes.Equal(output, []byte(expected)) {
		t.Fatalf("Expected %q, got %q", expected, output)
	}
}
