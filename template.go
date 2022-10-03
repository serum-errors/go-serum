package serum

import (
	"strings"
)

/*
This file contains a very small templating system.
It's used in some of the convenience constructors for serum errors.
It's a golang-special; none of this is part of the serum standards.

We do not reuse the stdlib `text/template` system because a simpler solution is adequate.
We **only** need basic string interpolation; function calls, etc, are unnecessary,
and indeed even undesired (since these would result in error message strings and details
attachments diverging from each other, which is hard to imagine being a good thing).
(Also: the `text/template` system doesn't allow controlling the implementation of lookups,
which would result in problematic efficiency barriers for our situation.)

There are no errors.
Questionably-formed template strings just produce questionably-formed output strings.
Lookups for values that aren't defined just produce the template syntax markers wrapped around the lookup key.
This is a principled choice: when in the middle of error handling, the last thing you want to do is
to get stuck debugging an explosive error from a templating system; malformed text is better than nothing.

The interpolation system uses the `[][2]string` system instead of `map[string]string`:
this is handy because we already use that elsewhere for its order-preservation properties;
it's also, coincidentally, much more allocation-friendly than map creation is.
It may scale poorly to large volumes of data, but in practice,
we expect this system to be used on data quantities where linear scan is much cheaper than even a single additional GC burden incidence.
*/

type parsed struct {
	literal string
	interp  string
}

func parse(s string) (result []parsed) {
	for {
		start := strings.Index(s, "{{")
		if start < 0 {
			result = append(result, parsed{literal: s})
			return
		}
		end := strings.Index(s[start+2:], "}}")
		if end < 0 {
			result = append(result, parsed{literal: s})
			return
		}
		if start > 0 {
			result = append(result, parsed{literal: s[0:start]})
		}
		if end > 0 {
			result = append(result, parsed{interp: strings.TrimSpace(s[start+2 : start+2+end])})
		}
		if end == 0 { // edgecase: if we found "{{}}", treat it like a literal.
			result = append(result, parsed{literal: s[start : start+4]})
		}
		s = s[start+end+4:]
		if s == "" {
			return
		}
	}
}

// Composes a string.
// Linear lookup into table.  Not expected to be used with large data.
// Does not bother to reuse strings.Builder buffers; possible target for future optimization, at the cost of synchronization.
func interpolate(ps []parsed, table [][2]string) string {
	var sb strings.Builder
	for _, p := range ps {
		if p.literal != "" {
			sb.WriteString(p.literal)
		}
		if p.interp != "" {
			var match bool
			for _, row := range table {
				if row[0] == p.interp {
					match = true
					sb.WriteString(row[1])
					break
				}
			}
			if !match {
				sb.WriteString("{{")
				sb.WriteString(p.interp)
				sb.WriteString("}}")
			}
		}
	}
	return sb.String()
}
