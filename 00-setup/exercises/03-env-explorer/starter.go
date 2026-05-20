// Package envexplorer wraps `go env <KEY>` lookups.
package envexplorer

import "fmt"

// GoEnv returns the value of a Go environment variable (e.g. "GOOS", "GOPATH").
// Equivalent to running `go env <key>` and returning its trimmed stdout.
// TODO: implement.
func GoEnv(key string) (string, error) {
	return "", fmt.Errorf("not implemented")
}
