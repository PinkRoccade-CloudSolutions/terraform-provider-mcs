package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &NetworkPoolResource{}
	_ resource.ResourceWithImportState = &NetworkPoolResource{}
)

type NetworkPoolResource struct {
	client *apiclient.Client
}

type NetworkPoolResourceModel struct {
	Id          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Network     types.String `tfsdk:"network"`
	Description types.String `tfsdk:"description"`
	Type        types.String `tfsdk:"type"`
	Enabled     types.Bool   `tfsdk:"enabled"`
}

type networkPoolResourceAPIModel struct {
	Id          string  `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Network     *string `json:"network,omitempty"`
	Description *string `json:"description,omitempty"`
	Type        *string `json:"type,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
}

func NewNetworkPoolResource() resource.Resource {
	return &NetworkPoolResource{}
}

func (r *NetworkPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_networkpool"
}

func (r *NetworkPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	optionalString := func(desc string, validators ...validator.String) schema.StringAttribute {
		return schema.StringAttribute{
			Optional:      true,
			Computed:      true,
			Description:   desc,
			Validators:    validators,
			PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	}

	resp.Schema = schema.Schema{
		Description: "Manages a network pool in MCS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the network pool.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name":        optionalString("Name of the network pool.", stringvalidator.LengthAtMost(255)),
			"network":     optionalString("Pool network in CIDR format, e.g. 10.0.0.0/8.", stringvalidator.LengthAtMost(255)),
			"description": optionalString("Description to help end users pick the right pool.", stringvalidator.LengthAtMost(2048)),
			"type":        optionalString("Pool type: lan, wan or transit.", stringvalidator.OneOf("lan", "wan", "transit", "")),
			"enabled": schema.BoolAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Whether the pool is usable by end users.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *NetworkPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NetworkPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NetworkPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp networkPoolResourceAPIModel
	if err := r.client.Post(ctx, "/api/networking/networkpools/", buildNetworkPoolAPIRequest(&plan), &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating network pool", err.Error())
		return
	}

	mapNetworkPoolToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NetworkPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp networkPoolResourceAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/networking/networkpools/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading network pool", err.Error())
		return
	}

	mapNetworkPoolToState(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *NetworkPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state NetworkPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp networkPoolResourceAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/networking/networkpools/%s/", state.Id.ValueString()), buildNetworkPoolAPIRequest(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating network pool", err.Error())
		return
	}

	mapNetworkPoolToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *NetworkPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NetworkPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/networking/networkpools/%s/", state.Id.ValueString()))
	if err != nil && !apiclient.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting network pool", err.Error())
	}
}

func (r *NetworkPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildNetworkPoolAPIRequest(plan *NetworkPoolResourceModel) networkPoolResourceAPIModel {
	return networkPoolResourceAPIModel{
		Name:        stringPtr(plan.Name),
		Network:     stringPtr(plan.Network),
		Description: stringPtr(plan.Description),
		Type:        stringPtr(plan.Type),
		Enabled:     boolPtr(plan.Enabled),
	}
}

func mapNetworkPoolToState(state *NetworkPoolResourceModel, api *networkPoolResourceAPIModel) {
	str := func(s *string) types.String {
		if s == nil {
			return types.StringValue("")
		}
		return types.StringValue(*s)
	}
	state.Id = types.StringValue(api.Id)
	state.Name = str(api.Name)
	state.Network = str(api.Network)
	state.Description = str(api.Description)
	state.Type = str(api.Type)
	state.Enabled = types.BoolValue(api.Enabled != nil && *api.Enabled)
}
