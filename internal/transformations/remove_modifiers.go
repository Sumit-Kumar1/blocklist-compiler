package transformations

import (
	"context"
	"slices"
	"strings"
)

// RemoveModifiers strips the modifiers a DNS filter cannot honour ($third-party,
// $3p, $document, $doc, $all, $popup, $network), keeping the rule itself rather
// than letting Validate drop it.
//
// It mirrors @adguard/hostlist-compiler's RemoveModifiers transformation.
type RemoveModifiers struct{}

func (r RemoveModifiers) Apply(ctx context.Context, data []byte) ([]byte, error) {
	return mapLines(ctx, r.Name(), data, func(line string) (string, bool) {
		if line == "" || isComment(line) {
			return line, true
		}

		pattern, modifiers, found := cutModifiers(line)
		if !found {
			return line, true
		}

		kept := slices.DeleteFunc(strings.Split(modifiers, ","), func(modifier string) bool {
			return slices.Contains(unsupportedModifiers, strings.TrimSpace(modifier))
		})

		if len(kept) == 0 {
			return pattern, true
		}

		return pattern + "$" + strings.Join(kept, ","), true
	})
}

func (r RemoveModifiers) Name() string {
	return "RemoveModifiers"
}
