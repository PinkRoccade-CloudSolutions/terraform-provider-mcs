package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const secureIngressFirewallBasePath = "/api/secureingress/firewall/"

var (
	_ resource.Resource                = &SecureIngressFirewallResource{}
	_ resource.ResourceWithImportState = &SecureIngressFirewallResource{}
)

type SecureIngressFirewallResource struct {
	client *apiclient.Client
}

type SecureIngressFirewallResourceModel struct {
	Id                 types.String         `tfsdk:"id"`
	Name               types.String         `tfsdk:"name"`
	Description        types.String         `tfsdk:"description"`
	FilterJSON         jsontypes.Normalized `tfsdk:"filter_json"`
	Customer           types.String         `tfsdk:"customer"`
	CreatedAtTimestamp types.String         `tfsdk:"created_at_timestamp"`
	UpdatedAtTimestamp types.String         `tfsdk:"updated_at_timestamp"`
	CreatedByUser      types.Int64          `tfsdk:"created_by_user"`
	UpdatedByUser      types.Int64          `tfsdk:"updated_by_user"`
}

// secureIngressFirewallRequest is the writable subset of the XDPFirewall schema.
// filter_json is a string in the spec, so the JSON document is sent as an encoded string.
type secureIngressFirewallRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
	FilterJSON  string  `json:"filter_json"`
	Customer    string  `json:"customer"`
}

type secureIngressFirewallAPIModel struct {
	Id                 string  `json:"id"`
	Name               string  `json:"name"`
	Description        *string `json:"description"`
	FilterJSON         string  `json:"filter_json"`
	Customer           string  `json:"customer"`
	CreatedAtTimestamp string  `json:"created_at_timestamp"`
	UpdatedAtTimestamp string  `json:"updated_at_timestamp"`
	CreatedByUser      *int64  `json:"created_by_user"`
	UpdatedByUser      *int64  `json:"updated_by_user"`
}

func NewSecureIngressFirewallResource() resource.Resource {
	return &SecureIngressFirewallResource{}
}

func (r *SecureIngressFirewallResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secureingress_firewall"
}

func (r *SecureIngressFirewallResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a secure ingress (XDP) firewall filter. Referenced by `mcs_ingress_cluster.firewall`.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the firewall.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the firewall (max 200 characters).",
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 200)},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Description: "Description of the firewall (max 200 characters).",
				Validators:  []validator.String{stringvalidator.LengthAtMost(200)},
			},
			"filter_json": schema.StringAttribute{
				Required:    true,
				CustomType:  jsontypes.NormalizedType{},
				Description: "Firewall filter definition as a JSON document (use `jsonencode(...)`). Formatting differences are ignored.",
			},
			"customer": schema.StringAttribute{
				Required:      true,
				Description:   "Customer identifier. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"created_at_timestamp": schema.StringAttribute{
				Computed:      true,
				Description:   "Time when the firewall was created.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at_timestamp": schema.StringAttribute{
				Computed:    true,
				Description: "Time when the firewall was last updated.",
			},
			"created_by_user": schema.Int64Attribute{
				Computed:      true,
				Description:   "ID of the user who created the firewall.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"updated_by_user": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the user who last updated the firewall.",
			},
		},
	}
}

func (r *SecureIngressFirewallResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *apiclient.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func buildSecureIngressFirewallRequest(plan *SecureIngressFirewallResourceModel) secureIngressFirewallRequest {
	return secureIngressFirewallRequest{
		Name:        plan.Name.ValueString(),
		Description: stringPtr(plan.Description),
		FilterJSON:  plan.FilterJSON.ValueString(),
		Customer:    plan.Customer.ValueString(),
	}
}

// secureIngressOptionalString maps a nullable API string onto an Optional (non-Computed) attribute.
// An empty string is kept null when the attribute was null before, so an unset value does not diff.
func secureIngressOptionalString(v *string, prior types.String) types.String {
	if v == nil || (*v == "" && prior.IsNull()) {
		return types.StringNull()
	}
	return types.StringValue(*v)
}

func mapSecureIngressFirewallToState(m *SecureIngressFirewallResourceModel, a *secureIngressFirewallAPIModel) {
	m.Id = types.StringValue(a.Id)
	m.Name = types.StringValue(a.Name)
	m.Description = secureIngressOptionalString(a.Description, m.Description)
	m.FilterJSON = jsontypes.NewNormalizedValue(a.FilterJSON)
	m.Customer = types.StringValue(a.Customer)
	m.CreatedAtTimestamp = types.StringValue(a.CreatedAtTimestamp)
	m.UpdatedAtTimestamp = types.StringValue(a.UpdatedAtTimestamp)
	m.CreatedByUser = types.Int64PointerValue(a.CreatedByUser)
	m.UpdatedByUser = types.Int64PointerValue(a.UpdatedByUser)
}

func (r *SecureIngressFirewallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecureIngressFirewallResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp secureIngressFirewallAPIModel
	if err := r.client.Post(ctx, secureIngressFirewallBasePath, buildSecureIngressFirewallRequest(&plan), &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating secure ingress firewall", err.Error())
		return
	}

	mapSecureIngressFirewallToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SecureIngressFirewallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecureIngressFirewallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp secureIngressFirewallAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("%s%s/", secureIngressFirewallBasePath, state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading secure ingress firewall", err.Error())
		return
	}

	mapSecureIngressFirewallToState(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SecureIngressFirewallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SecureIngressFirewallResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp secureIngressFirewallAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("%s%s/", secureIngressFirewallBasePath, state.Id.ValueString()), buildSecureIngressFirewallRequest(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating secure ingress firewall", err.Error())
		return
	}

	mapSecureIngressFirewallToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SecureIngressFirewallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecureIngressFirewallResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("%s%s/", secureIngressFirewallBasePath, state.Id.ValueString()))
	if err != nil && !apiclient.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting secure ingress firewall", err.Error())
	}
}

func (r *SecureIngressFirewallResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
