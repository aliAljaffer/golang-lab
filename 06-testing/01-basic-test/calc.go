// 01-basic-test — the minimum viable Go test.
//
// One file with the function under test, one `_test.go` file next to it,
// `go test ./...` discovers it. No framework, no annotations, no config.
package calc

// Add returns a + b. Trivial on purpose — the point of this example is the
// surrounding `_test.go`, not the function.
func Add(a, b int) int {
	return a + b
}

// Sub returns a - b.
func Sub(a, b int) int {
	return a - b
}
