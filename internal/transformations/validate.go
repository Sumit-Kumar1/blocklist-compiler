package transformations

import (
	"context"
	"slices"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// unsupportedModifiers cannot be honoured by a DNS filter, so rules carrying
// them are dropped rather than silently misapplied.
var unsupportedModifiers = []string{
	"third-party", "3p", "document", "doc", "all", "popup", "network",
}

// limitingModifiers narrow a rule's scope, rescuing it from the public-suffix check.
var limitingModifiers = []string{"denyallow", "client", "ctag", "badfilter"}

// Validate removes rules that are dangerous or meaningless for DNS filtering:
// TLD-wide blocks, IP-address rules, cosmetic rules, dotless hosts and rules
// carrying modifiers a DNS filter cannot honour.
//
// It mirrors @adguard/hostlist-compiler's Validate transformation.
type Validate struct{}

func (v Validate) Apply(ctx context.Context, data []byte) ([]byte, error) {
	return mapLines(ctx, v.Name(), data, func(line string) (string, bool) {
		if line == "" || isComment(line) {
			return line, true
		}

		// Hosts entries are validated by their own format, not adblock syntax.
		if _, isHosts := hostsDomains(line); isHosts {
			return line, true
		}

		return line, isSafeRule(line)
	})
}

// isSafeRule reports whether a single non-hosts line may stay in the list.
func isSafeRule(line string) bool {
	if isCosmetic(line) {
		return false
	}

	// A standalone /regex/ rule carries no host to validate.
	if strings.HasPrefix(line, "/") && strings.HasSuffix(line, "/") {
		return true
	}

	pattern, modifiers, _ := strings.Cut(line, "$")
	if hasUnsupportedModifier(modifiers) {
		return false
	}

	return isSafeHost(strings.TrimPrefix(pattern, "@@"), modifiers)
}

// isSafeHost validates the host a rule pattern targets, if it names one at all.
func isSafeHost(pattern, modifiers string) bool {
	// A bare IP is too broad however it is written.
	if isIPLiteral(pattern) {
		return false
	}

	// Only host-anchored rules carry a host to validate. Substring patterns such
	// as "-728-x-90-" are matched against URLs and are left alone.
	anchored, isHostRule := strings.CutPrefix(pattern, "||")
	if !isHostRule {
		return true
	}

	// "||adidas-outlet*" is a prefix pattern rather than a complete host, so
	// there is no hostname to check.
	host, terminated := strings.CutSuffix(anchored, "^")
	if !terminated {
		return true
	}

	// Bare IPs and dotless hosts are too broad to be useful.
	if isIPLiteral(host) || !strings.Contains(host, ".") || !isRuleHost(host) {
		return false
	}

	// "||*.org^" and "||bet.ar^" take out every domain under a public suffix,
	// unless a modifier narrows the rule back down.
	return hasLimitingModifier(modifiers) || !isPublicSuffix(strings.TrimPrefix(host, "*."))
}

// hasLimitingModifier reports whether a modifier narrows an otherwise
// overly broad rule, which upstream accepts as making it safe.
func hasLimitingModifier(modifiers string) bool {
	for modifier := range strings.SplitSeq(modifiers, ",") {
		name, _, _ := strings.Cut(modifier, "=")

		if slices.Contains(limitingModifiers, strings.TrimSpace(name)) {
			return true
		}
	}

	return false
}

// isPublicSuffix reports whether host is itself an ICANN registry suffix such as
// "org", "co.uk" or "bet.ar", rather than a registrable domain.
//
// Only the ICANN section counts. Privately operated suffixes such as "us.kg",
// "co.com" and "giize.com" host ordinary subdomains that blocklists legitimately
// target, and upstream keeps rules for them.
func isPublicSuffix(host string) bool {
	if host == "" {
		return false
	}

	suffix, icann := publicsuffix.PublicSuffix(host)

	return icann && suffix == host
}

func hasUnsupportedModifier(modifiers string) bool {
	if modifiers == "" {
		return false
	}

	for modifier := range strings.SplitSeq(modifiers, ",") {
		name, _, _ := strings.Cut(modifier, "=")
		name = strings.TrimPrefix(strings.TrimSpace(name), "~")

		if slices.Contains(unsupportedModifiers, name) {
			return true
		}
	}

	return false
}

func (v Validate) Name() string {
	return "Validate"
}
