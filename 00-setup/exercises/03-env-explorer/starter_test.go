//go:build exercise

package envexplorer

import (
	"runtime"
	"testing"
)

func TestGoEnv_GOOS(t *testing.T) {
	got, err := GoEnv("GOOS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != runtime.GOOS {
		t.Errorf("GOOS: got %q, want %q", got, runtime.GOOS)
	}
}

func TestGoEnv_GOARCH(t *testing.T) {
	got, err := GoEnv("GOARCH")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != runtime.GOARCH {
		t.Errorf("GOARCH: got %q, want %q", got, runtime.GOARCH)
	}
}

func TestGoEnv_GOPATH_NonEmpty(t *testing.T) {
	got, err := GoEnv("GOPATH")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got == "" {
		t.Error("GOPATH should not be empty")
	}
}

func TestGoEnv_NoTrailingWhitespace(t *testing.T) {
	got, err := GoEnv("GOOS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) > 0 && (got[len(got)-1] == '\n' || got[len(got)-1] == ' ') {
		t.Errorf("GOOS value has trailing whitespace: %q", got)
	}
}
