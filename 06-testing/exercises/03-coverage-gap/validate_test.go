//go:build exercise

package validate

import (
	"errors"
	"testing"
)

// validConfig returns a config that passes Validate. Use it as a base in your
// tests so you only have to mutate the one field you're testing.
func validConfig() Config {
	return Config{
		Name:    "svc",
		Port:    8080,
		Tags:    []string{"http", "api"},
		Env:     "staging",
		Timeout: 30,
	}
}

// One starter test. The exercise: add tests until `go test -cover` reports 100%.
func TestValidate_HappyPath(t *testing.T) {
	if err := Validate(validConfig()); err != nil {
		t.Errorf("Validate(valid) = %v, want nil", err)
	}
}

func TestValidate_EmptyName(t *testing.T) {
	c := validConfig()
	c.Name = ""
	if err := Validate(c); !errors.Is(err, ErrEmptyName) {
		t.Errorf("Validate(empty name) = %v, want ErrEmptyName", err)
	}
}

// TODO: branches still uncovered (run `go test -cover ...` to confirm):
//   - name longer than 64 chars
//   - port < 1
//   - port > 65535
//   - port < 1024 AND env == "prod" (the cross-field rule — tricky)
//   - env == "" (unknown env via empty string)
//   - env == "qa" (unknown env via non-empty unknown)
//   - tag contains a space
//   - tag contains a tab
//   - tag contains a newline (probably already covered by ContainsAny — verify)
//   - timeout < 0
//
// Add one test per branch. Aim for `coverage: 100.0% of statements`.
