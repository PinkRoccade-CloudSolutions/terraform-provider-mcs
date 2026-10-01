package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &IPPoolResource{}
	_ resource.ResourceWithImportState = &IPPoolResource{}
)

type IPPoolResource struct {
	client *apiclient.Client
}

type IPPoolResourceModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Subnet   types.String `tfsdk:"subnet"`
	Type     types.String `tfsdk:"type"`
	Customer types.String `tfsdk:"customer"`
	TotalIPs types.String `tfsdk:"total_ips"`
	FreeIPs  types.String `tfsdk:"free_ips"`
}

type ippoolResourceAPIModel struct {
	Id       string      `json:"id,omitempty"`
	Name     string      `json:"name"`
	Subnet   string      `json:"subnet"`
	Type     *string     `json:"type,omitempty"`
	Customer *string     `json:"customer,omitempty"`
	TotalIPs ippoolCount `json:"total_ips,omitempty"`
	FreeIPs  ippoolCount `json:"free_ips,omitempty"`
}

func NewIPPoolResource() resource.Resource {
	return &IPPoolResource{}
}

func (r *IPPoolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ippool"
}

func (r *IPPoolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages an IP pool in MCS.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the IP pool.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the IP pool.",
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 100)},
			},
			"subnet": schema.StringAttribute{
				Required:      true,
				Description:   "Subnet of the IP pool, e.g. 203.0.113.0/24. Changing this forces a new resource.",
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 50)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"type": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Pool type: nat, vip or loadbalancer.",
				Validators:    []validator.String{stringvalidator.OneOf("nat", "vip", "loadbalancer", "")},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"customer": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Customer owning the IP pool.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"total_ips": schema.StringAttribute{
				Computed:      true,
				Description:   "Total number of IP addresses in the pool.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"free_ips": schema.StringAttribute{
				Computed:    true,
				Description: "Number of free IP addresses in the pool.",
			},
		},
	}
}

func (r *IPPoolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *IPPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IPPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp ippoolResourceAPIModel
	if err := r.client.Post(ctx, "/api/networking/ippools/", buildIPPoolAPIRequest(&plan), &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating IP pool", err.Error())
		return
	}

	mapIPPoolToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IPPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IPPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp ippoolResourceAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/networking/ippools/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading IP pool", err.Error())
		return
	}

	mapIPPoolToState(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IPPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state IPPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp ippoolResourceAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/networking/ippools/%s/", state.Id.ValueString()), buildIPPoolAPIRequest(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating IP pool", err.Error())
		return
	}

	mapIPPoolToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IPPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IPPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/networking/ippools/%s/", state.Id.ValueString()))
	if err != nil && !apiclient.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting IP pool", err.Error())
	}
}

func (r *IPPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func buildIPPoolAPIRequest(plan *IPPoolResourceModel) ippoolResourceAPIModel {
	return ippoolResourceAPIModel{
		Name:     plan.Name.ValueString(),
		Subnet:   plan.Subnet.ValueString(),
		Type:     stringPtr(plan.Type),
		Customer: stringPtr(plan.Customer),
	}
}

func mapIPPoolToState(state *IPPoolResourceModel, api *ippoolResourceAPIModel) {
	state.Id = types.StringValue(api.Id)
	state.Name = types.StringValue(api.Name)
	state.Subnet = types.StringValue(api.Subnet)
	state.Type = types.StringValue("")
	if api.Type != nil {
		state.Type = types.StringValue(*api.Type)
	}
	state.Customer = types.StringPointerValue(api.Customer)
	state.TotalIPs = types.StringValue(string(api.TotalIPs))
	state.FreeIPs = types.StringValue(string(api.FreeIPs))
}
