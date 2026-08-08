package transformations

import (
	"net/netip"
	"strings"
)

// cosmeticMarkers identify element-hiding and scriptlet rules, which are
// meaningless for a DNS blocklist.
var cosmeticMarkers = []string{"##", "#@#", "#?#", "#$#", "#%#", "#@$#"}

// isComment reports whether a line is an adblock (!) or hosts (#) comment.
func isComment(line string) bool {
	return strings.HasPrefix(line, "!") || strings.HasPrefix(line, "#")
}

// hostsDomains splits a hosts-format line ("<ip> domain [domain...]") into its
// domains. ok is false when the line is not hosts format.
func hostsDomains(line string) (domains []string, ok bool) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return nil, false
	}

	if _, err := netip.ParseAddr(fields[0]); err != nil {
		return nil, false
	}

	for _, field := range fields[1:] {
		// A trailing "# comment" ends the entry.
		if strings.HasPrefix(field, "#") {
			break
		}

		domains = append(domains, field)
	}

	return domains, len(domains) > 0
}

// isCosmetic reports whether the rule hides elements rather than blocking a host.
func isCosmetic(line string) bool {
	for _, marker := range cosmeticMarkers {
		if strings.Contains(line, marker) {
			return true
		}
	}

	return false
}

// blockedDomain returns the domain of a plain "||domain^" blocking rule.
// ok is false for allow rules, rules carrying modifiers, and anything else.
func blockedDomain(line string) (domain string, ok bool) {
	rest, found := strings.CutPrefix(line, "||")
	if !found {
		return "", false
	}

	rest, found = strings.CutSuffix(rest, "^")
	if !found || rest == "" {
		return "", false
	}

	// A modifier or separator anywhere means this is not a plain host rule.
	if strings.ContainsAny(rest, "^$/*") {
		return "", false
	}

	return rest, true
}

// isPlainDomain reports whether a line is a bare hostname such as "example.com",
// with no adblock syntax and no surrounding whitespace. Patterns that merely
// look domain-ish ("*outlet.", "-1080x140-") are rejected, matching upstream.
func isPlainDomain(line string) bool {
	if line == "" || line != strings.TrimSpace(line) {
		return false
	}

	if isIPLiteral(line) {
		return true
	}

	if !strings.Contains(line, ".") {
		return false
	}

	for label := range strings.SplitSeq(line, ".") {
		if !isHostLabel(label) {
			return false
		}
	}

	return true
}

// isRuleHost reports whether host is well-formed for a "||host^" rule. Unlike a
// plain domain, an adblock pattern may carry "*" wildcards in any label.
func isRuleHost(host string) bool {
	for label := range strings.SplitSeq(host, ".") {
		if label == "" {
			return false
		}

		for _, r := range label {
			switch {
			case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '*':
			default:
				return false
			}
		}
	}

	return true
}

// isHostLabel reports whether s is a valid DNS label.
func isHostLabel(s string) bool {
	if s == "" || strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return false
	}

	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-':
		default:
			return false
		}
	}

	return true
}

// isIPLiteral reports whether a line is a bare IP address.
func isIPLiteral(line string) bool {
	_, err := netip.ParseAddr(line)

	return err == nil
}

// coversSubdomain reports whether parent is a proper DNS ancestor of domain.
func coversSubdomain(parent, domain string) bool {
	return len(domain) > len(parent) && strings.HasSuffix(domain, "."+parent)
}
