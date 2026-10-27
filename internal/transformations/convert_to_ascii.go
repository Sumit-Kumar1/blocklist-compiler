package transformations

import (
	"context"
	"strings"
	"unicode"

	"golang.org/x/net/idna"
)

// ConvertToASCII rewrites internationalized domains as punycode, so
// "||пример.рф^" becomes "||xn--e1afmkfd.xn--p1ai^". Without it Validate drops
// such rules as malformed.
//
// It mirrors @adguard/hostlist-compiler's ConvertToAscii transformation.
type ConvertToASCII struct{}

func (c ConvertToASCII) Apply(ctx context.Context, data []byte) ([]byte, error) {
	return mapLines(ctx, c.Name(), data, func(line string) (string, bool) {
		if isASCII(line) || isComment(line) {
			return line, true
		}

		return punycodeHosts(line), true
	})
}

// punycodeHosts converts every run of hostname characters holding non-ASCII
// text, leaving the surrounding rule syntax untouched.
func punycodeHosts(line string) string {
	var buf strings.Builder

	start := -1

	for i, r := range line {
		if isHostRune(r) {
			if start < 0 {
				start = i
			}

			continue
		}

		if start >= 0 {
			buf.WriteString(toASCII(line[start:i]))

			start = -1
		}

		buf.WriteRune(r)
	}

	if start >= 0 {
		buf.WriteString(toASCII(line[start:]))
	}

	return buf.String()
}

func isHostRune(r rune) bool {
	return r == '.' || r == '-' || unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r)
}

// toASCII punycodes host, returning it unchanged when it cannot be converted.
func toASCII(host string) string {
	if isASCII(host) {
		return host
	}

	converted, err := idna.ToASCII(host)
	if err != nil {
		return host
	}

	return converted
}

func isASCII(s string) bool {
	return strings.IndexFunc(s, func(r rune) bool { return r > unicode.MaxASCII }) < 0
}

func (c ConvertToASCII) Name() string {
	return "ConvertToAscii"
}
