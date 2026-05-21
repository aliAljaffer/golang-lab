// 02-table-driven — parametrize one test over many inputs with a slice of structs.
//
// Idiomatic Go: instead of writing N near-identical `Test*` functions, declare
// a slice of `{name, input, want}` and loop over it. One assertion, many cases.
package strs

// Reverse returns s with its runes reversed. Works for unicode.
func Reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < j; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
