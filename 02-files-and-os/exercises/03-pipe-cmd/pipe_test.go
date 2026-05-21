//go:build exercise && !windows

package pipecmd

import (
	"strings"
	"testing"
)

func TestPipe_Single(t *testing.T) {
	got, err := Pipe(strings.NewReader("hello\n"), []string{"cat"})
	if err != nil {
		t.Fatalf("Pipe: %v", err)
	}
	if string(got) != "hello\n" {
		t.Errorf("got %q, want %q", got, "hello\n")
	}
}

func TestPipe_TwoStage(t *testing.T) {
	// cat | tr a-z A-Z
	got, err := Pipe(strings.NewReader("hello\n"), []string{"cat"}, []string{"tr", "a-z", "A-Z"})
	if err != nil {
		t.Fatalf("Pipe: %v", err)
	}
	if string(got) != "HELLO\n" {
		t.Errorf("got %q, want %q", got, "HELLO\n")
	}
}

func TestPipe_ThreeStage(t *testing.T) {
	// printf-style input | sort | wc -l  =>  count of lines
	in := strings.NewReader("c\na\nb\n")
	got, err := Pipe(in, []string{"sort"}, []string{"wc", "-l"})
	if err != nil {
		t.Fatalf("Pipe: %v", err)
	}
	// wc -l output is whitespace-padded on macOS, trim before compare.
	if strings.TrimSpace(string(got)) != "3" {
		t.Errorf("got %q, want trimmed %q", got, "3")
	}
}

func TestPipe_PropagatesFailure(t *testing.T) {
	// `false` exits non-zero — must propagate as an error.
	_, err := Pipe(strings.NewReader(""), []string{"false"})
	if err == nil {
		t.Fatal("expected error from `false`, got nil")
	}
}

func TestPipe_NoCommandsIsError(t *testing.T) {
	if _, err := Pipe(strings.NewReader("")); err == nil {
		t.Fatal("expected error for empty pipeline, got nil")
	}
}
