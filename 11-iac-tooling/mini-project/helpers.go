package main

import (
	"bytes"
	"errors"
	"os"
	"text/template"
)

// RenderTemplate renders `tmpl` against `vars` using text/template.
// A missing key is an error (`missingkey=error`) — silent substitution of
// "<no value>" into a file Terraform manages is a security smell.
func RenderTemplate(tmpl string, vars map[string]string) (string, error) {
	// TODO: parse tmpl with template.New("file").Option("missingkey=error").
	// TODO: execute against vars into a bytes.Buffer.
	// TODO: return buf.String(), nil — or wrap parse/exec error.
	_ = template.New
	_ = bytes.Buffer{}
	return "", errors.New("RenderTemplate not implemented")
}

// WriteTemplatedFile renders the template and writes the result to `path`.
// Returns the rendered content (what landed on disk).
func WriteTemplatedFile(path, tmpl string, vars map[string]string) (string, error) {
	// TODO: content, err := RenderTemplate(tmpl, vars).
	// TODO: if err: return "", err.
	// TODO: os.WriteFile(path, []byte(content), 0o644).
	// TODO: return content, nil — or the write error.
	_ = path
	_ = tmpl
	_ = vars
	return "", errors.New("WriteTemplatedFile not implemented")
}

// ReadTemplatedFile reads the file at `path`.
// Returns os.ErrNotExist (wrapped) if the file is gone — that's how the
// resource detects drift.
func ReadTemplatedFile(path string) (string, error) {
	// TODO: b, err := os.ReadFile(path) — return ("", err) on error
	// TODO: (os.ErrNotExist is already wrapped by os.ReadFile, so just propagate).
	// TODO: return string(b), nil.
	_ = path
	return "", errors.New("ReadTemplatedFile not implemented")
}

// DeleteTemplatedFile removes the file. Missing file is not an error —
// Terraform may call Delete on a resource whose file was already removed
// manually, and a re-run of `terraform destroy` must not fail.
func DeleteTemplatedFile(path string) error {
	// TODO: err := os.Remove(path).
	// TODO: if errors.Is(err, os.ErrNotExist): return nil.
	// TODO: return err.
	_ = path
	_ = os.Remove
	return errors.New("DeleteTemplatedFile not implemented")
}
