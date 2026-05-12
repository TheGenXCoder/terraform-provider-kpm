package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/TheGenXCoder/terraform-provider-kpm/internal/client"
)

var _ resource.Resource = &SecretResource{}

// SecretResource manages a KPM secret via the kpm_secret Terraform resource.
type SecretResource struct {
	client client.AgentKMSClient
}

// SetClient is used by tests to inject a mock client.
func (r *SecretResource) SetClient(c client.AgentKMSClient) { r.client = c }

// NewSecretResource returns the constructor used by the provider.
func NewSecretResource() resource.Resource { return &SecretResource{} }

type secretModel struct {
	Path        types.String `tfsdk:"path"`
	Value       types.String `tfsdk:"value"`
	Type        types.String `tfsdk:"type"`
	Description types.String `tfsdk:"description"`
	Tags        types.List   `tfsdk:"tags"`
}

func (r *SecretResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secret"
}

func (r *SecretResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a secret stored in AgentKMS via KPM.",
		Attributes: map[string]schema.Attribute{
			"path": schema.StringAttribute{
				Required:    true,
				Description: "Secret path in service/name format (e.g. 'db/password').",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Secret value. Stored sensitive in Terraform state.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Secret type: generic (default), api-token, ssh-key, connection-string, jwt, password.",
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Human-readable description of the secret.",
			},
			"tags": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				Description: "Tags for filtering (e.g. ['prod', 'db']).",
			},
		},
	}
}

func (r *SecretResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(client.AgentKMSClient)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("expected client.AgentKMSClient, got %T", req.ProviderData))
		return
	}
	r.client = c
}

func (r *SecretResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan secretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := toStringSlice(ctx, plan.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.WriteSecret(ctx,
		plan.Path.ValueString(),
		plan.Value.ValueString(),
		tags,
		plan.Description.ValueString(),
		plan.Type.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("Error creating secret", err.Error())
		return
	}

	if plan.Type.IsNull() || plan.Type.IsUnknown() {
		plan.Type = types.StringValue("generic")
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SecretResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state secretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	value, err := r.client.GetSecret(ctx, state.Path.ValueString())
	if err != nil {
		// Secret no longer exists — remove from state.
		resp.State.RemoveResource(ctx)
		return
	}
	state.Value = types.StringValue(value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SecretResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan secretModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	tags := toStringSlice(ctx, plan.Tags, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.WriteSecret(ctx,
		plan.Path.ValueString(),
		plan.Value.ValueString(),
		tags,
		plan.Description.ValueString(),
		plan.Type.ValueString(),
	); err != nil {
		resp.Diagnostics.AddError("Error updating secret", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SecretResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state secretModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteSecret(ctx, state.Path.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error deleting secret", err.Error())
	}
}

func toStringSlice(ctx context.Context, list types.List, diags *diag.Diagnostics) []string {
	if list.IsNull() || list.IsUnknown() {
		return nil
	}
	var out []string
	diags.Append(list.ElementsAs(ctx, &out, false)...)
	return out
}
