// Package fileinfods is exercise 11-iac/01 — implement a Terraform data source.
//
// A data source is the read-only sibling of a resource. The user declares
//
//	data "fileops_file_info" "x" { path = "/etc/passwd" }
//
// and Terraform calls Read to populate `exists`, `size`, etc. There's no
// Create/Update/Delete — just Schema + Read.
//
// Your job: implement the data source so that:
//
//  1. Its TypeName is `<provider>_file_info`.
//  2. Its Schema declares: `path` (Required), `exists` (Computed bool),
//     `size` (Computed int64).
//  3. Its Read handler populates `exists` and `size` from the actual file
//     on disk.
//
// The bulk of the testable surface is the pure helper `ReadFileInfo` —
// implement that first, then wire it into the framework Read handler.
package fileinfods

import (
	"context"
	"errors"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FileInfo is the pure-Go shape of the data source's output state.
// Tests target this, not the framework's types.Bool etc.
type FileInfo struct {
	Path   string
	Exists bool
	Size   int64
}

// ReadFileInfo statss `path` and returns the FileInfo. Missing files set
// Exists=false (NOT an error — a non-existent file is a valid query
// result). Other errors (permission denied, etc.) are returned.
func ReadFileInfo(path string) (FileInfo, error) {
	// TODO: stat the file, then map the result into a FileInfo. The docstring
	//   above pins the policy: ErrNotExist is *not* an error here (the data
	//   source answers "does it exist?" — false is a valid answer); other
	//   errors propagate. Use errors.Is to match the sentinel across wrappers.
	_ = errors.Is
	_ = os.Stat
	return FileInfo{}, errors.New("ReadFileInfo not implemented")
}

// FileInfoModel mirrors the schema for Get/Set.
type FileInfoModel struct {
	Path   types.String `tfsdk:"path"`
	Exists types.Bool   `tfsdk:"exists"`
	Size   types.Int64  `tfsdk:"size"`
}

// FileInfoDS satisfies datasource.DataSource.
type FileInfoDS struct{}

func NewFileInfoDS() datasource.DataSource { return &FileInfoDS{} }

func (d *FileInfoDS) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	// TODO: set the type name. Convention is "<provider>_<thing>" — the
	//   provider half is on req, the thing half is what the package doc names.
	_ = req
	_ = resp
}

func (d *FileInfoDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// TODO: declare the three attributes from the package doc. Decide which
	//   is the user input (Required) and which are populated by Read
	//   (Computed) — the input/output split here defines the data source.
	//   Attribute types live in dsschema.{String,Bool,Int64}Attribute.
	_ = resp
	_ = dsschema.Schema{}
}

func (d *FileInfoDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// TODO: the Read flow is: pull the user input out of req.Config into a
	//   FileInfoModel, call ReadFileInfo, then write the result back into
	//   resp.State. Each framework call returns diagnostics — append them and
	//   bail early when HasError, otherwise a later step will panic on
	//   half-populated state. Wrap Go primitives in types.{Bool,Int64}Value
	//   before assignment.
	_ = ctx
	_ = req
	_ = resp
}
