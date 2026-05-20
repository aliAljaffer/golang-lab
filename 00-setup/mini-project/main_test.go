//go:build exercise

package main

import (
	"testing"
)

func TestParseGoMod_Basic(t *testing.T) {
	content := []byte(`module github.com/example/demo

go 1.22
`)
	got, err := parseGoMod(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Module != "github.com/example/demo" {
		t.Errorf("module: got %q, want %q", got.Module, "github.com/example/demo")
	}
	if got.Go != "1.22" {
		t.Errorf("go version: got %q, want %q", got.Go, "1.22")
	}
	if got.Deps != 0 {
		t.Errorf("deps: got %d, want 0", got.Deps)
	}
}

func TestParseGoMod_WithDeps(t *testing.T) {
	content := []byte(`module github.com/example/demo

go 1.22

require (
	github.com/spf13/cobra v1.8.0
	github.com/stretchr/testify v1.8.4
)
`)
	got, err := parseGoMod(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Deps != 2 {
		t.Errorf("deps: got %d, want 2", got.Deps)
	}
}

func TestParseGoMod_SingleLineRequire(t *testing.T) {
	content := []byte(`module example.com/x

go 1.22

require github.com/google/uuid v1.6.0
`)
	got, err := parseGoMod(content)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Deps != 1 {
		t.Errorf("deps: got %d, want 1", got.Deps)
	}
}
