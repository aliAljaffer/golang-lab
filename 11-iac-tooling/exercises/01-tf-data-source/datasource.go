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
	// TODO: stat the path with os.Stat.
	// TODO: if errors.Is(err, os.ErrNotExist): return FileInfo{Path: path, Exists: false}, nil
	// TODO: if err != nil: return FileInfo{}, err
	// TODO: return FileInfo{Path: path, Exists: true, Size: st.Size()}, nil
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
	// TODO: resp.TypeName = req.ProviderTypeName + "_file_info"
	_ = req
	_ = resp
}

func (d *FileInfoDS) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	// TODO: resp.Schema = dsschema.Schema{
	// TODO:     Attributes: map[string]dsschema.Attribute{
	// TODO:         "path":   dsschema.StringAttribute{Required: true},
	// TODO:         "exists": dsschema.BoolAttribute{Computed: true},
	// TODO:         "size":   dsschema.Int64Attribute{Computed: true},
	// TODO:     },
	// TODO: }
	_ = resp
	_ = dsschema.Schema{}
}

func (d *FileInfoDS) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// TODO: var data FileInfoModel
	// TODO: resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	// TODO: if resp.Diagnostics.HasError() { return }
	// TODO: info, err := ReadFileInfo(data.Path.ValueString())
	// TODO: if err != nil { resp.Diagnostics.AddError("stat failed", err.Error()); return }
	// TODO: data.Exists = types.BoolValue(info.Exists)
	// TODO: data.Size = types.Int64Value(info.Size)
	// TODO: resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	_ = ctx
	_ = req
	_ = resp
}
