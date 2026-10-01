package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &PublicIPAddressResource{}

type PublicIPAddressResource struct {
	client *apiclient.Client
}

type PublicIPAddressResourceModel struct {
	Id          types.String `tfsdk:"id"`
	IPAddress   types.String `tfsdk:"ip_address"`
	Pool        types.String `tfsdk:"pool"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
	Type        types.String `tfsdk:"type"`
	Customer    types.String `tfsdk:"customer"`
}

type publicIPAddressAPIModel struct {
	Id          string  `json:"id,omitempty"`
	IPAddress   *string `json:"ip_address,omitempty"`
	Pool        *string `json:"pool,omitempty"`
	Description string  `json:"description"`
	Status      string  `json:"status,omitempty"`
	Type        *string `json:"type,omitempty"`
	Customer    *string `json:"customer,omitempty"`
}

func NewPublicIPAddressResource() resource.Resource {
	return &PublicIPAddressResource{}
}

func (r *PublicIPAddressResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ip_address"
}

func (r *PublicIPAddressResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a public IP address in MCS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the public IP address.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"ip_address": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "The public IP address. Assigned by the API when not set.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"pool": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of the IP pool this address belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString(""),
				Description: "Description of the public IP address.",
				Validators:  []validator.String{stringvalidator.LengthAtMost(255)},
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Allocation status: available, assigned or reserved (read-only).",
			},
			"type": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Type: nat, vip, loadbalancer or secureingress.",
				Validators:    []validator.String{stringvalidator.OneOf("nat", "vip", "loadbalancer", "secureingress", "")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"customer": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Customer identifier.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *PublicIPAddressResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PublicIPAddressResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PublicIPAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := buildPublicIPAPIRequest(&plan)

	var apiResp publicIPAddressAPIModel
	err := r.client.Post(ctx, "/api/networking/publicipaddresss/", apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating public IP address", err.Error())
		return
	}

	mapPublicIPToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PublicIPAddressResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PublicIPAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp publicIPAddressAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/networking/publicipaddresss/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading public IP address", err.Error())
		return
	}

	mapPublicIPToState(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *PublicIPAddressResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state PublicIPAddressResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiReq := buildPublicIPAPIRequest(&plan)

	var apiResp publicIPAddressAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/networking/publicipaddresss/%s/", state.Id.ValueString()), apiReq, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating public IP address", err.Error())
		return
	}

	mapPublicIPToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *PublicIPAddressResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PublicIPAddressResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/networking/publicipaddresss/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting public IP address", err.Error())
	}
}

func buildPublicIPAPIRequest(plan *PublicIPAddressResourceModel) publicIPAddressAPIModel {
	return publicIPAddressAPIModel{
		IPAddress:   stringPtr(plan.IPAddress),
		Pool:        stringPtr(plan.Pool),
		Description: plan.Description.ValueString(),
		Type:        stringPtr(plan.Type),
		Customer:    stringPtr(plan.Customer),
	}
}

func mapPublicIPToState(state *PublicIPAddressResourceModel, api *publicIPAddressAPIModel) {
	state.Id = types.StringValue(api.Id)
	state.IPAddress = types.StringPointerValue(api.IPAddress)
	state.Pool = types.StringPointerValue(api.Pool)
	state.Description = types.StringValue(api.Description)
	state.Status = types.StringValue(api.Status)
	state.Type = types.StringValue("")
	if api.Type != nil {
		state.Type = types.StringValue(*api.Type)
	}
	state.Customer = types.StringPointerValue(api.Customer)
}
