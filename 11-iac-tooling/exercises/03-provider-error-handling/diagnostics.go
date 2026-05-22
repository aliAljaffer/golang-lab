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
	// TODO: if err == nil: return nil.
	// TODO: var diags diag.Diagnostics
	// TODO: switch {
	// TODO: case errors.Is(err, os.ErrPermission):
	// TODO:     diags.AddError("permission denied",
	// TODO:         "cannot write to "+path+
	// TODO:             ": check file mode (chmod) and ownership (chown), "+
	// TODO:             "or run terraform with sufficient privileges.")
	// TODO: case errors.Is(err, os.ErrNotExist):
	// TODO:     diags.AddError("parent directory does not exist",
	// TODO:         "cannot write to "+path+
	// TODO:             ": the parent directory does not exist. "+
	// TODO:             "Create it first (e.g., with a local-exec or another fileops_dir resource).")
	// TODO: default:
	// TODO:     diags.AddError("write failed", "write to "+path+" failed: "+err.Error())
	// TODO: }
	// TODO: return diags
	_ = errors.Is
	_ = os.ErrPermission
	return nil
}
