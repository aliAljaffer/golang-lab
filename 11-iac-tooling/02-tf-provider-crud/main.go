// 02-tf-provider-crud — the same `echo_value` resource, but with the full
// CRUD lifecycle fleshed out.
//
// What's different from 01:
//   - the Read handler actually reads state and re-emits it (the canonical
//     refresh pattern — if you had a remote API, you'd call it here and
//     overwrite state with the truth).
//   - the Update handler is shown explicitly (Plan-in, State-out).
//   - the resource now has a `length` *computed* attribute that the
//     provider derives from `input` — demonstrates "computed by the
//     provider, not by the user."
//   - a validator on `input` enforces non-empty strings via
//     `stringvalidator.LengthAtLeast(1)`.
//   - the resource has an `id` that's a function of `input` (a fake hash)
//     — so a different `input` triggers a replace, same `input` is a noop.
//
// Run the provider locally:
//
//	go build -o terraform-provider-echo
//	BIN_PATH=~/.terraform.d/plugins/registry.terraform.io/examples/echo/0.2.0/$(go env GOOS)_$(go env GOARCH)
//	mkdir -p "$BIN_PATH" && mv terraform-provider-echo "$BIN_PATH/"
//
// Then in a scratch dir, write:
//
//	terraform {
//	  required_providers { echo = { source = "registry.terraform.io/examples/echo", version = "0.2.0" } }
//	}
//	resource "echo_value" "x" { input = "hello" }
//	output "out" { value = echo_value.x.output }
//
// `terraform init && terraform apply` — observe Create. Change `input`,
// apply again — observe Replace. `terraform destroy` — observe Delete.
package main

import (
	"context"
	"crypto/sha1"
	"encoding/hex"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	providerschema "github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
)

type echoProvider struct{}

func newProvider() provider.Provider { return &echoProvider{} }

func (p *echoProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "echo"
	resp.Version = "0.2.0"
}

func (p *echoProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
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

type echoModel struct {
	ID     types.String `tfsdk:"id"`
	Input  types.String `tfsdk:"input"`
	Output types.String `tfsdk:"output"`
	Length types.Int64  `tfsdk:"length"`
}

type echoResource struct{}

func newEchoResource() resource.Resource { return &echoResource{} }

func (r *echoResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_value"
}

func (r *echoResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		Description: "A useless resource that echoes its input.",
		Attributes: map[string]resourceschema.Attribute{
			"id": resourceschema.StringAttribute{
				Computed: true,
				// UseStateForUnknown prevents the framework from marking `id`
				// as unknown during updates — without it every plan would
				// show `id` as "(known after apply)".
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"input": resourceschema.StringAttribute{
				Required:    true,
				Description: "The string to echo.",
				Validators: []validator.String{
					stringvalidator.LengthAtLeast(1),
				},
				// RequiresReplace: changing input recreates the resource.
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"output": resourceschema.StringAttribute{Computed: true},
			"length": resourceschema.Int64Attribute{Computed: true},
		},
	}
}

func (r *echoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// TODO: read the plan into an echoModel.
	// TODO: compute Output / Length / ID from Input.
	// TODO: write the populated model to state.
	//
	// Hint: req.Plan.Get(ctx, &data) returns Diagnostics; append them and
	// return early if HasError(). Mirror with resp.State.Set(ctx, &data).
	var data echoModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	populate(&data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *echoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Canonical refresh pattern. For a remote-backed resource, you'd:
	//   1. read state into the model
	//   2. call the remote API with the model's ID
	//   3. overwrite the model with what the API returned
	//   4. write back to state (or call resp.State.RemoveResource if 404)
	// For this echo resource there's no remote — we just re-emit state.
	var data echoModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *echoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// TODO: same shape as Create, but the source is req.Plan and the sink
	// TODO: is resp.State. RequiresReplace on `input` means Update is only
	// TODO: reached for non-replacement-causing attribute changes — for this
	// TODO: resource, that's currently nothing, but it's good practice to
	// TODO: have the handler exist anyway.
	var data echoModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}
	populate(&data)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *echoResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
	// Nothing to delete in a remote system. The framework removes the
	// resource from state automatically on Delete success.
}

// populate is the pure-Go "remote API" for this resource. Real providers
// would call AWS / GCP / a database here.
func populate(d *echoModel) {
	in := d.Input.ValueString()
	d.Output = types.StringValue(in)
	d.Length = types.Int64Value(int64(len(in)))
	sum := sha1.Sum([]byte(in))
	d.ID = types.StringValue(hex.EncodeToString(sum[:8]))
}

func main() {
	_ = providerserver.Serve(context.Background(), newProvider, providerserver.ServeOpts{
		Address: "registry.terraform.io/examples/echo",
	})
}
