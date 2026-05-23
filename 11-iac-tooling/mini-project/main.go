// tf-provider-fileops — a Terraform provider that manages templated local files.
//
// Resource: `fileops_templated_file`
//
//	resource "fileops_templated_file" "x" {
//	  path     = "/tmp/hello.txt"
//	  template = "hello, {{.name}}"
//	  vars     = { name = "world" }
//	}
//
// On apply: renders the template against vars, writes to path.
// On refresh: re-reads the file, updates state.content.
// On replace (path change): destroys + recreates.
// On in-place update (template/vars change): re-renders + overwrites.
// On destroy: removes the file (idempotent — missing file is fine).
//
// The CRUD handlers delegate to four pure helpers:
//
//	RenderTemplate(tmpl, vars) (string, error)
//	WriteTemplatedFile(path, tmpl, vars) (content string, err error)
//	ReadTemplatedFile(path) (content string, err error) -- os.ErrNotExist on 404
//	DeleteTemplatedFile(path) error -- idempotent
//
// Splitting the I/O into pure helpers means most of the testing surface is
// reachable WITHOUT a real `terraform` binary. Acceptance tests (gated on
// TF_ACC=1) cover the full plan/apply lifecycle end-to-end.
//
// Production stance baked in:
//   - templates use `missingkey=error`. A missing var fails plan/apply
//     loudly instead of writing the string "<no value>" to disk. That's
//     the right default for an IaC tool — silent value substitution is a
//     security smell.
//   - DeleteTemplatedFile ignores os.ErrNotExist. Terraform's destroy is
//     re-runnable; the second destroy must not fail because the file is
//     already gone.
package main

import (
	"context"
	"errors"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// --- Provider ---------------------------------------------------------------

type fileopsProvider struct{}

func NewProvider() provider.Provider { return &fileopsProvider{} }

func (p *fileopsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "fileops"
	resp.Version = "0.1.0"
}

func (p *fileopsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = providerschema.Schema{}
}

func (p *fileopsProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *fileopsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{newTemplatedFile}
}

func (p *fileopsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// --- Resource ---------------------------------------------------------------

type templatedFileModel struct {
	Path     types.String `tfsdk:"path"`
	Template types.String `tfsdk:"template"`
	Vars     types.Map    `tfsdk:"vars"`
	Content  types.String `tfsdk:"content"`
	ID       types.String `tfsdk:"id"`
}

type templatedFile struct{}

func newTemplatedFile() resource.Resource { return &templatedFile{} }

func (r *templatedFile) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_templated_file"
}

func (r *templatedFile) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		Description: "A file on the local filesystem whose contents come from a Go template + vars.",
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"path": resourceschema.StringAttribute{
				Required:      true,
				Description:   "Where to write the file. Changing this destroys and recreates.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"template": resourceschema.StringAttribute{
				Required:    true,
				Description: "Go text/template body. {{.var}} placeholders pull from `vars`.",
			},
			"vars": resourceschema.MapAttribute{
				Required:    true,
				ElementType: types.StringType,
				Description: "Variables passed to the template.",
			},
			"content": resourceschema.StringAttribute{
				Computed:    true,
				Description: "The rendered file contents (what was actually written).",
			},
		},
	}
}

// extractVars converts a types.Map (from state/plan) into a plain Go map.
func extractVars(ctx context.Context, m types.Map) (map[string]string, error) {
	out := map[string]string{}
	if m.IsNull() || m.IsUnknown() {
		return out, nil
	}
	diags := m.ElementsAs(ctx, &out, false)
	if diags.HasError() {
		return nil, errors.New(diags.Errors()[0].Detail())
	}
	return out, nil
}

func (r *templatedFile) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TODO: Create flow: pull the user's config from req.Plan, do the write,
	//   then persist Computed attributes (content, id) back to resp.State.
	//   Two things that will bite you if forgotten:
	//     - vars is a types.Map; extractVars unboxes it for the helper.
	//     - the id attribute is what Terraform uses to identify the resource
	//       across runs; pick something stable from the user input.
	//   Every framework call returns diagnostics — append + bail on HasError.
	_ = ctx
	_ = req
	_ = resp
}

func (r *templatedFile) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// TODO: Read is the refresh step. Pull current state, re-stat the file,
	//   reconcile. The interesting case is "file is gone" — Terraform's word
	//   for that is *drift*, and there's a specific resp.State call that tells
	//   the framework "drop this from state so the next plan recreates it".
	//   Distinguish that from other errors (which should surface to the user).
	_ = ctx
	_ = req
	_ = resp
}

func (r *templatedFile) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: Update is Create's twin — same write path, but the id was
	//   already established and must round-trip from prior state, not be
	//   re-derived from the plan. (path changes go through Replace, not here,
	//   because of the RequiresReplace plan modifier on path.)
	_ = ctx
	_ = req
	_ = resp
}

func (r *templatedFile) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// TODO: pull the path out of state and ask the helper to delete it. The
	//   idempotency policy (missing file is fine) lives in the helper, not
	//   here — don't duplicate it.
	_ = ctx
	_ = req
	_ = resp
	_ = os.ErrNotExist
}

func main() {
	_ = providerserver.Serve(context.Background(), NewProvider, providerserver.ServeOpts{
		Address: "registry.terraform.io/examples/fileops",
	})
}
