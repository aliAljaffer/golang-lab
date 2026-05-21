// 02-envdump — list env vars matching a glob; optionally unset them.
package envdump

// Pair is one env var entry.
type Pair struct {
	Key   string
	Value string
}

// Match returns the subset of `env` whose keys match `pattern`.
// pattern is a shell-style glob (path/filepath.Match semantics): `*` matches
// any run of chars, `?` matches one char, character classes with `[abc]`.
// Match against the *key* only, not the value.
//
// `env` is a slice of "KEY=VALUE" strings (the same format as os.Environ()).
// Order of the result must match the order of `env`.
func Match(env []string, pattern string) ([]Pair, error) {
	// TODO: implement using path/filepath.Match.
	return nil, nil
}

// Unsetter abstracts os.Unsetenv so tests can verify which keys would be cleared
// without actually mutating the process environment.
type Unsetter interface {
	Unsetenv(key string) error
}

// UnsetMatching calls unsetter.Unsetenv(key) for each Pair returned by Match.
// Returns the list of keys it attempted to unset (in order), and the first
// error encountered (or nil).
func UnsetMatching(env []string, pattern string, unsetter Unsetter) ([]string, error) {
	// TODO: implement.
	return nil, nil
}
