// 03-subtests — `t.Run(name, func(t *testing.T) {...})` makes each case its
// own runnable, filterable, parallelizable subtest.
package strs

import "strings"

// Repeat returns s concatenated n times. n < 0 is treated as 0.
func Repeat(s string, n int) string {
	if n < 0 {
		n = 0
	}
	return strings.Repeat(s, n)
}
