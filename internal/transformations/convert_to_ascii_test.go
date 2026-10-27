package transformations

import (
	"bytes"
	"context"
	"testing"
)

func TestConvertToASCII(t *testing.T) {
	input := `! пример comment
||пример.рф^
||*.пример.рф^$important
@@||bücher.de^
0.0.0.0 пример.рф
||ascii.com^
`
	expected := `! пример comment
||xn--e1afmkfd.xn--p1ai^
||*.xn--e1afmkfd.xn--p1ai^$important
@@||xn--bcher-kva.de^
0.0.0.0 xn--e1afmkfd.xn--p1ai
||ascii.com^
`

	output, err := ConvertToASCII{}.Apply(context.Background(), []byte(input))
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !bytes.Equal(output, []byte(expected)) {
		t.Fatalf("Expected %q, got %q", expected, output)
	}
}
