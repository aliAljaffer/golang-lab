// 01-tf-provider-skeleton — the smallest Terraform provider that compiles.
//
// A Terraform provider is a Go binary that speaks gRPC to the `terraform`
// CLI. When you write `resource "echo_value" "x" { ... }`, terraform looks
// up the `echo` provider plugin, exec()s it, and talks to it over a unix
// socket. The plugin's job is to answer Plan/Apply/Refresh by implementing
// the resource lifecycle (CRUD).
//
// We use `terraform-plugin-framework` (the modern SDK). The older
// `terraform-plugin-sdk/v2` is still around but new providers should use
// the framework — it has cleaner ergonomics and proper context support.
//
// What this skeleton has:
//   - one provider (`echo`), no provider-level config
//   - one resource (`echo_value`) with `input` (required) and `output`
//     (computed) — Create copies input → output; Read/Update/Delete are
//     no-ops because this resource has no remote state.
//
// What you'll write here:
//   - resource Schema (attributes + their constraints)
//   - the Create handler (read plan, set state)
//   - Update mirroring Create
//
// To actually run this against `terraform`:
//
//	go build -o terraform-provider-echo
//	mkdir -p ~/.terraform.d/plugins/registry.terraform.io/examples/echo/0.1.0/$(go env GOOS)_$(go env GOARCH)
//	mv terraform-provider-echo ~/.terraform.d/plugins/registry.terraform.io/examples/echo/0.1.0/$(go env GOOS)_$(go env GOARCH)/
//	# then write a .tf file using `provider "echo" {}` and `resource "echo_value" "x" { input = "hi" }`
//
// See example 02 for real CRUD with state.
package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// echoProvider is the provider. It satisfies `provider.Provider`.
type echoProvider struct{}

func newProvider() provider.Provider { return &echoProvider{} }

func (p *echoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "echo"
	resp.Version = "0.1.0"
}

func (p *echoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	// No provider-level config (no API key, no endpoint). Empty schema is fine.
	resp.Schema = providerschema.Schema{}
}

func (p *echoProvider) Configure(_ context.Context, _ provider.ConfigureRequest, _ *provider.ConfigureResponse) {
}

func (p *echoProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{newEchoResource}
}

func (p *echoProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

// echoModel is the shape of the resource's state in Terraform.
// Each field maps to one schema attribute via the `tfsdk` tag.
type echoModel struct {
	ID     types.String `tfsdk:"id"`
	Input  types.String `tfsdk:"input"`
	Output types.String `tfsdk:"output"`
}

type echoResource struct{}

func newEchoResource() resource.Resource { return &echoResource{} }

func (r *echoResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	// `<provider_type_name>_<resource_suffix>` → `echo_value`
	resp.TypeName = req.ProviderTypeName + "_value"
}

func (r *echoResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	// TODO: declare the resource's attributes.
	// TODO: resp.Schema = resourceschema.Schema{
	// TODO:     Attributes: map[string]resourceschema.Attribute{
	// TODO:         "id":     resourceschema.StringAttribute{Computed: true},
	// TODO:         "input":  resourceschema.StringAttribute{Required: true},
	// TODO:         "output": resourceschema.StringAttribute{Computed: true},
	// TODO:     },
	// TODO: }
	_ = resp
	_ = resourceschema.Schema{}
}

func (r *echoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TODO: var data echoModel
	// TODO: resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	// TODO: if resp.Diagnostics.HasError() { return }
	// TODO: data.Output = data.Input
	// TODO: data.ID = types.StringValue("echo")
	// TODO: resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	_ = ctx
	_ = req
	_ = resp
	_ = echoModel{}
	_ = types.StringValue
}

// Read is a no-op for this stateless resource. Real providers Refresh here.
func (r *echoResource) Read(_ context.Context, _ resource.ReadRequest, _ *resource.ReadResponse) {
}

func (r *echoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Mirror Create — the plan IS the new state.
	var data echoModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	data.Output = data.Input
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *echoResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// No remote system → nothing to clean up.
}

func main() {
	_ = providerserver.Serve(context.Background(), newProvider, providerserver.ServeOpts{
		Address: "registry.terraform.io/examples/echo",
	})
}
