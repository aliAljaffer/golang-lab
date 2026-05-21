// 08-fuzz — Go's built-in fuzzer (Go 1.18+). Generates random inputs derived
// from your seed corpus, runs them through a target function, and reports any
// input that panics or violates an invariant.
//
// This example intentionally ships a buggy ParseInt so the fuzzer has
// something to find.
package parse

import (
	"errors"
)

var ErrEmpty = errors.New("empty input")

// ParseInt parses a base-10 signed integer. It has a bug — see if the fuzzer
// can spot it. (Hint: think about strings that contain only a sign.)
func ParseInt(s string) (int, error) {
	if len(s) == 0 {
		return 0, ErrEmpty
	}
	neg := false
	i := 0
	if s[0] == '-' {
		neg = true
		i = 1
	}
	// BUG: if s == "-", the loop body never runs and we return 0, nil.
	// A correct impl would check that i < len(s) before falling through.
	n := 0
	for ; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, errors.New("non-digit")
		}
		n = n*10 + int(c-'0')
	}
	if neg {
		n = -n
	}
	return n, nil
}
