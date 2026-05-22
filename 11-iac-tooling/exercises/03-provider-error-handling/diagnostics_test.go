//go:build exercise

package provdiag

import (
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

func TestWriteFileDiagnostics_NilErrReturnsEmpty(t *testing.T) {
	d := WriteFileDiagnostics("/anywhere", nil)
	if d.HasError() {
		t.Fatalf("expected no diags, got %d", d.ErrorsCount())
	}
}

func TestWriteFileDiagnostics_PermissionDeniedHintsChmod(t *testing.T) {
	d := WriteFileDiagnostics("/root/locked", os.ErrPermission)
	if !d.HasError() {
		t.Fatal("expected error diag for permission denied")
	}
	errs := d.Errors()
	if len(errs) != 1 {
		t.Fatalf("got %d diagnostics, want 1", len(errs))
	}
	got := errs[0]
	if !strings.Contains(strings.ToLower(got.Summary()), "permission") {
		t.Errorf("summary %q should mention permission", got.Summary())
	}
	detail := strings.ToLower(got.Detail())
	// Detail must point the user at something they can do.
	if !strings.Contains(detail, "chmod") && !strings.Contains(detail, "chown") && !strings.Contains(detail, "privile") {
		t.Errorf("detail %q should hint at chmod/chown/privileges", got.Detail())
	}
	if !strings.Contains(got.Detail(), "/root/locked") {
		t.Errorf("detail %q should include the path", got.Detail())
	}
}

func TestWriteFileDiagnostics_MissingParentHintsCreateDir(t *testing.T) {
	d := WriteFileDiagnostics("/nope/sub/file", os.ErrNotExist)
	if !d.HasError() {
		t.Fatal("expected error diag for missing parent dir")
	}
	got := d.Errors()[0]
	if !strings.Contains(strings.ToLower(got.Summary()), "directory") &&
		!strings.Contains(strings.ToLower(got.Summary()), "parent") {
		t.Errorf("summary %q should mention parent/directory", got.Summary())
	}
	detail := strings.ToLower(got.Detail())
	if !strings.Contains(detail, "create") && !strings.Contains(detail, "mkdir") {
		t.Errorf("detail %q should hint at creating the dir", got.Detail())
	}
}

func TestWriteFileDiagnostics_GenericErrorIncludesUnderlying(t *testing.T) {
	raw := errors.New("disk full")
	d := WriteFileDiagnostics("/var/lib/app/data", raw)
	if !d.HasError() {
		t.Fatal("expected error diag")
	}
	got := d.Errors()[0]
	if !strings.Contains(got.Detail(), "disk full") {
		t.Errorf("detail %q should include the underlying error string", got.Detail())
	}
	if !strings.Contains(got.Detail(), "/var/lib/app/data") {
		t.Errorf("detail %q should include the path", got.Detail())
	}
}

func TestWriteFileDiagnostics_AllSeveritiesAreError(t *testing.T) {
	// Warnings would let a user merrily apply and then wonder why nothing
	// happened. Anything that prevented a write is an error.
	cases := []error{os.ErrPermission, os.ErrNotExist, errors.New("disk full")}
	for _, e := range cases {
		d := WriteFileDiagnostics("/some/path", e)
		for _, dd := range d {
			if dd.Severity() != diag.SeverityError {
				t.Errorf("err %v produced severity %v, want SeverityError", e, dd.Severity())
			}
		}
	}
}
