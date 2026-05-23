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
	// TODO: render via text/template. The docstring above pins the policy
	//   that matters — missing keys must error, not silently emit "<no value>".
	//   Look up how to set that on a *template.Template before parsing.
	_ = template.New
	_ = bytes.Buffer{}
	return "", errors.New("RenderTemplate not implemented")
}

// WriteTemplatedFile renders the template and writes the result to `path`.
// Returns the rendered content (what landed on disk).
func WriteTemplatedFile(path, tmpl string, vars map[string]string) (string, error) {
	// TODO: render first, then write. The returned content is exactly what
	//   landed on disk — Terraform stores it under .content for drift detection,
	//   so do not return the un-rendered template by accident.
	_ = path
	_ = tmpl
	_ = vars
	return "", errors.New("WriteTemplatedFile not implemented")
}

// ReadTemplatedFile reads the file at `path`.
// Returns os.ErrNotExist (wrapped) if the file is gone — that's how the
// resource detects drift.
func ReadTemplatedFile(path string) (string, error) {
	// TODO: read the file. The contract above demands that os.ErrNotExist
	//   survives errors.Is — the resource's Read uses that to detect drift.
	//   Pick a stdlib call that preserves it (or wrap deliberately).
	_ = path
	return "", errors.New("ReadTemplatedFile not implemented")
}

// DeleteTemplatedFile removes the file. Missing file is not an error —
// Terraform may call Delete on a resource whose file was already removed
// manually, and a re-run of `terraform destroy` must not fail.
func DeleteTemplatedFile(path string) error {
	// TODO: delete the file, but make the "already gone" case a success —
	//   the docstring above explains why (re-runnable destroy). Every other
	//   error propagates.
	_ = path
	_ = os.Remove
	return errors.New("DeleteTemplatedFile not implemented")
}
