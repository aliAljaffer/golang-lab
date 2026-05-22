//go:build exercise

package main

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// --- RenderTemplate ---------------------------------------------------------

func TestRenderTemplate_HappyPath(t *testing.T) {
	got, err := RenderTemplate("hello, {{.name}}", map[string]string{"name": "world"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "hello, world" {
		t.Fatalf("got %q", got)
	}
}

func TestRenderTemplate_MissingVarIsError(t *testing.T) {
	// missingkey=error is the production-grade default. Silent <no value>
	// substitution in an IaC tool would let typos ship to disk.
	_, err := RenderTemplate("hi {{.missing}}", map[string]string{})
	if err == nil {
		t.Fatal("expected error for missing var, got nil")
	}
}

func TestRenderTemplate_InvalidSyntax(t *testing.T) {
	_, err := RenderTemplate("{{.bad", map[string]string{})
	if err == nil {
		t.Fatal("expected parse error, got nil")
	}
}

// --- WriteTemplatedFile -----------------------------------------------------

func TestWriteTemplatedFile_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	content, err := WriteTemplatedFile(path, "answer={{.n}}", map[string]string{"n": "42"})
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if content != "answer=42" {
		t.Fatalf("content=%q", content)
	}
	onDisk, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readback: %v", err)
	}
	if string(onDisk) != "answer=42" {
		t.Fatalf("on-disk=%q", string(onDisk))
	}
}

func TestWriteTemplatedFile_OverwritesExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(path, []byte("OLD"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteTemplatedFile(path, "NEW", map[string]string{}); err != nil {
		t.Fatalf("err: %v", err)
	}
	b, _ := os.ReadFile(path)
	if string(b) != "NEW" {
		t.Fatalf("got %q want NEW", string(b))
	}
}

// --- ReadTemplatedFile ------------------------------------------------------

func TestReadTemplatedFile_HappyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "in.txt")
	if err := os.WriteFile(path, []byte("contents"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := ReadTemplatedFile(path)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if got != "contents" {
		t.Fatalf("got %q", got)
	}
}

func TestReadTemplatedFile_NotFoundReturnsErrNotExist(t *testing.T) {
	// The resource's Read handler relies on errors.Is(err, os.ErrNotExist)
	// to drop the resource from state on drift. Regression-catch the
	// "wrapping" so future-me doesn't replace os.ReadFile with something
	// that loses the sentinel.
	_, err := ReadTemplatedFile(filepath.Join(t.TempDir(), "missing.txt"))
	if !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("got %v, want os.ErrNotExist", err)
	}
}

// --- DeleteTemplatedFile ----------------------------------------------------

func TestDeleteTemplatedFile_Removes(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doomed.txt")
	if err := os.WriteFile(path, []byte("bye"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := DeleteTemplatedFile(path); err != nil {
		t.Fatalf("err: %v", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("file still exists or unexpected stat err: %v", err)
	}
}

func TestDeleteTemplatedFile_MissingIsIdempotent(t *testing.T) {
	// terraform destroy is re-runnable. Second destroy after manual rm must succeed.
	if err := DeleteTemplatedFile(filepath.Join(t.TempDir(), "ghost.txt")); err != nil {
		t.Fatalf("expected nil for missing file, got %v", err)
	}
}

// --- Provider plumbing ------------------------------------------------------

// TestProviderSchemaCompiles is a smoke test: building a provider instance
// and reading its metadata should never error. Catches "forgot to register
// the resource" mistakes.
func TestProviderSchemaCompiles(t *testing.T) {
	p := NewProvider()
	if p == nil {
		t.Fatal("NewProvider returned nil")
	}
	// Indirect way to verify the resource is wired in: the factory list
	// must be non-empty.
	res := p.Resources(t.Context())
	if len(res) == 0 {
		t.Fatal("expected at least one resource factory")
	}
	r := res[0]()
	if r == nil {
		t.Fatal("resource factory returned nil")
	}
}

// --- Acceptance test stub ---------------------------------------------------
//
// A full acceptance test would import
// github.com/hashicorp/terraform-plugin-testing/helper/resource and drive
// `terraform plan` + `apply` + `destroy` against a real terraform binary.
// That's gated on TF_ACC=1 by convention — without it, the test must skip
// (and not fail).
//
// Run with:
//
//	TF_ACC=1 go test -tags=exercise -run TestAcc -v ./11-iac-tooling/mini-project/
//
// Implementation is left as an extension exercise — see the README's
// "extending to a real acceptance test" section.
func TestAccTemplatedFile_SkipsWithoutTF_ACC(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run acceptance tests against a real terraform binary")
	}
	// In a real impl this would call resource.UnitTest or resource.Test.
	// For the scaffold, just assert the env var is correctly recognized.
	if !strings.EqualFold(os.Getenv("TF_ACC"), "1") {
		t.Fatalf("TF_ACC must be exactly 1")
	}
}
