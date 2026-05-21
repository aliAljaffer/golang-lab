//go:build exercise

package envdump

import "testing"

var sampleEnv = []string{
	"PATH=/usr/bin",
	"APP_HOST=localhost",
	"APP_PORT=8080",
	"HOME=/root",
	"APP_DEBUG=true",
}

func TestMatch_Prefix(t *testing.T) {
	got, err := Match(sampleEnv, "APP_*")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d, want 3 (APP_HOST, APP_PORT, APP_DEBUG)", len(got))
	}
	wantOrder := []string{"APP_HOST", "APP_PORT", "APP_DEBUG"}
	for i, p := range got {
		if p.Key != wantOrder[i] {
			t.Errorf("[%d] key=%q, want %q", i, p.Key, wantOrder[i])
		}
	}
}

func TestMatch_ExactKey(t *testing.T) {
	got, err := Match(sampleEnv, "HOME")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Value != "/root" {
		t.Errorf("got %+v, want one match HOME=/root", got)
	}
}

func TestMatch_NoMatches(t *testing.T) {
	got, err := Match(sampleEnv, "NOPE*")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Errorf("got %d, want 0", len(got))
	}
}

func TestMatch_InvalidPatternReturnsError(t *testing.T) {
	_, err := Match(sampleEnv, "[bad")
	if err == nil {
		t.Error("expected error for malformed glob, got nil")
	}
}

type fakeUnsetter struct {
	calls []string
}

func (f *fakeUnsetter) Unsetenv(key string) error {
	f.calls = append(f.calls, key)
	return nil
}

func TestUnsetMatching_UnsetsMatchedKeysOnly(t *testing.T) {
	u := &fakeUnsetter{}
	keys, err := UnsetMatching(sampleEnv, "APP_*", u)
	if err != nil {
		t.Fatal(err)
	}
	if len(keys) != 3 {
		t.Fatalf("returned %d keys, want 3", len(keys))
	}
	if len(u.calls) != 3 {
		t.Fatalf("unsetter called %d times, want 3", len(u.calls))
	}
	for i, want := range []string{"APP_HOST", "APP_PORT", "APP_DEBUG"} {
		if u.calls[i] != want {
			t.Errorf("call[%d] = %q, want %q", i, u.calls[i], want)
		}
	}
}
