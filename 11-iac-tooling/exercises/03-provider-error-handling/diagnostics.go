// Package provdiag is exercise 11-iac/03 — actionable provider diagnostics.
//
// The default failure mode of a Terraform provider is:
//
//	Error: write failed
//	  open /etc/myapp/config.yaml: permission denied
//
// True but useless. The user reads "permission denied" and asks "what now?"
//
// Your job: write `WriteFileDiagnostics(path string, err error) diag.Diagnostics`
// that translates raw `os` errors into actionable diagnostics:
//
//   - nil err          → empty diagnostics
//   - os.ErrPermission → Summary "permission denied",
//                        Detail mentions chmod / chown / running as root
//   - os.ErrNotExist   → Summary "parent directory does not exist",
//                        Detail mentions creating the dir
//                        (we hit this from os.WriteFile when the parent dir is missing —
//                        WriteFile only creates the file itself, not its parents)
//   - other            → Summary "write failed",
//                        Detail includes path + raw err string
//
// All diagnostics returned MUST have Severity = SeverityError.
package provdiag

import (
	"errors"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// WriteFileDiagnostics converts an os.WriteFile-style error into a slice
// of actionable Terraform diagnostics.
func WriteFileDiagnostics(path string, err error) diag.Diagnostics {
	// TODO: nothing-to-report case — decide what the caller sees when err is nil.
	// TODO: dispatch on the error kind (errors.Is against the os sentinels in
	//   the package doc; everything else is the generic branch) and AddError
	//   with the Summary/Detail the doc specifies. Detail must be *actionable* —
	//   the tests pin chmod/chown/privileges, create/mkdir, and that the path
	//   plus underlying error string appear in the generic case.
	_ = errors.Is
	_ = os.ErrPermission
	return nil
}
