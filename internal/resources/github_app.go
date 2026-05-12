package resources

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/TheGenXCoder/terraform-provider-kpm/internal/client"
)

var _ resource.Resource = &GithubAppResource{}

// GithubAppResource manages a GitHub App registration via the kpm_github_app Terraform resource.
type GithubAppResource struct {
	client client.AgentKMSClient
}

// NewGithubAppResource returns the constructor used by the provider.
func NewGithubAppResource() resource.Resource { return &GithubAppResource{} }

// SetClient injects a client — used by the provider and unit tests.
func (r *GithubAppResource) SetClient(c client.AgentKMSClient) { r.client = c }

type githubAppModel struct {
	Name             types.String `tfsdk:"name"`
	AppID            types.Int64  `tfsdk:"app_id"`
	InstallationID   types.Int64  `tfsdk:"installation_id"`
	PrivateKey       types.String `tfsdk:"private_key"`
	PrivateKeySHA256 types.String `tfsdk:"private_key_sha256"`
}

func (r *GithubAppResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_github_app"
}

func (r *GithubAppResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a GitHub App registration in AgentKMS via KPM.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name for the GitHub App registration.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"app_id": schema.Int64Attribute{
				Required:    true,
				Description: "GitHub App ID.",
			},
			"installation_id": schema.Int64Attribute{
				Required:    true,
				Description: "GitHub App Installation ID.",
			},
			"private_key": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "PEM-encoded GitHub App private key. Never stored in plain text; a SHA-256 fingerprint is stored instead.",
			},
			"private_key_sha256": schema.StringAttribute{
				Computed:    true,
				Description: "SHA-256 hex fingerprint of the private key, used for drift detection.",
			},
		},
	}
}

func (r *GithubAppResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GithubAppResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan githubAppModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.RegisterGithubApp(ctx, client.RegisterGithubAppRequest{
		Name:           plan.Name.ValueString(),
		AppID:          plan.AppID.ValueInt64(),
		InstallationID: plan.InstallationID.ValueInt64(),
		PrivateKeyPEM:  plan.PrivateKey.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error registering GitHub App", err.Error())
		return
	}

	plan.PrivateKeySHA256 = types.StringValue(sha256Hex(plan.PrivateKey.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GithubAppResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state githubAppModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := r.client.GetGithubApp(ctx, state.Name.ValueString())
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading GitHub App", err.Error())
		return
	}

	// Update server-authoritative fields; preserve private_key and private_key_sha256
	// because AgentKMS never returns the private key.
	state.AppID = types.Int64Value(app.AppID)
	state.InstallationID = types.Int64Value(app.InstallationID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *GithubAppResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan githubAppModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.RegisterGithubApp(ctx, client.RegisterGithubAppRequest{
		Name:           plan.Name.ValueString(),
		AppID:          plan.AppID.ValueInt64(),
		InstallationID: plan.InstallationID.ValueInt64(),
		PrivateKeyPEM:  plan.PrivateKey.ValueString(),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating GitHub App", err.Error())
		return
	}

	plan.PrivateKeySHA256 = types.StringValue(sha256Hex(plan.PrivateKey.ValueString()))
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *GithubAppResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state githubAppModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.RemoveGithubApp(ctx, state.Name.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error removing GitHub App", err.Error())
	}
}

// sha256Hex returns the lowercase hex-encoded SHA-256 digest of s.
func sha256Hex(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h)
}
