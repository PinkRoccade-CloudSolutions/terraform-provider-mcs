package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &FirewallServiceResource{}

type FirewallServiceResource struct {
	client *apiclient.Client
}

type FirewallServiceModel struct {
	Id           types.String `tfsdk:"id"`
	Domain       types.String `tfsdk:"domain"`
	Name         types.String `tfsdk:"name"`
	Uuid         types.String `tfsdk:"uuid"`
	Protocol     types.String `tfsdk:"protocol"`
	Comment      types.String `tfsdk:"comment"`
	TcpPortrange types.List   `tfsdk:"tcp_portrange"`
	UdpPortrange types.List   `tfsdk:"udp_portrange"`
	Used         types.Bool   `tfsdk:"used"`
}

type firewallServiceRequest struct {
	Name         string    `json:"name"`
	Protocol     string    `json:"protocol"`
	Comment      *string   `json:"comment,omitempty"`
	TcpPortrange *[]string `json:"tcp_portrange,omitempty"`
	UdpPortrange *[]string `json:"udp_portrange,omitempty"`
}

type firewallServiceAPI struct {
	Name         string   `json:"name"`
	Uuid         string   `json:"uuid"`
	Protocol     string   `json:"protocol"`
	Comment      *string  `json:"comment"`
	TcpPortrange []string `json:"tcp_portrange"`
	UdpPortrange []string `json:"udp_portrange"`
	Used         bool     `json:"used"`
}

func NewFirewallServiceResource() resource.Resource {
	return &FirewallServiceResource{}
}

func (r *FirewallServiceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_service"
}

func (r *FirewallServiceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a firewall service in the MCS API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain": schema.StringAttribute{
				Required:      true,
				Description:   "The domain this service belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the firewall service.",
			},
			"uuid": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID assigned by the firewall.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"protocol": schema.StringAttribute{
				Required:    true,
				Description: "Protocol type for the service (e.g. TCP/UDP/SCTP).",
			},
			"comment": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Comment for the service.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tcp_portrange": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of TCP port ranges. Removing the attribute clears the ranges.",
			},
			"udp_portrange": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of UDP port ranges. Removing the attribute clears the ranges.",
			},
			"used": schema.BoolAttribute{
				Computed:      true,
				Description:   "Whether the service is currently in use.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *FirewallServiceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type", fmt.Sprintf("Expected *apiclient.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func (r *FirewallServiceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallServiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := firewallServiceRequest{
		Name:         plan.Name.ValueString(),
		Protocol:     plan.Protocol.ValueString(),
		Comment:      stringPtr(plan.Comment),
		TcpPortrange: listElems[string](ctx, plan.TcpPortrange, &resp.Diagnostics),
		UdpPortrange: listElems[string](ctx, plan.UdpPortrange, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallServiceAPI
	path := fmt.Sprintf("/api/networking/domain/%s/services/", plan.Domain.ValueString())
	if err := r.client.Post(ctx, path, body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating firewall service", err.Error())
		return
	}

	mapFirewallServiceToState(ctx, &plan, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallServiceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallServiceAPI
	path := fmt.Sprintf("/api/networking/domain/%s/services/%s/", state.Domain.ValueString(), state.Name.ValueString())
	if err := r.client.Get(ctx, path, &apiResp); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall service", err.Error())
		return
	}

	mapFirewallServiceToState(ctx, &state, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallServiceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state FirewallServiceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := firewallServiceRequest{
		Name:         plan.Name.ValueString(),
		Protocol:     plan.Protocol.ValueString(),
		Comment:      stringPtr(plan.Comment),
		TcpPortrange: listElemsForUpdate[string](ctx, plan.TcpPortrange, state.TcpPortrange, &resp.Diagnostics),
		UdpPortrange: listElemsForUpdate[string](ctx, plan.UdpPortrange, state.UdpPortrange, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Address the service by its current (state) name so that renames work.
	var apiResp firewallServiceAPI
	path := fmt.Sprintf("/api/networking/domain/%s/services/%s/", state.Domain.ValueString(), state.Name.ValueString())
	if err := r.client.Patch(ctx, path, body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error updating firewall service", err.Error())
		return
	}

	mapFirewallServiceToState(ctx, &plan, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallServiceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallServiceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("/api/networking/domain/%s/services/%s/", state.Domain.ValueString(), state.Name.ValueString())
	if err := r.client.Delete(ctx, path); err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting firewall service", err.Error())
	}
}

func mapFirewallServiceToState(ctx context.Context, model *FirewallServiceModel, api *firewallServiceAPI, diags *diag.Diagnostics) {
	model.Id = types.StringValue(api.Uuid)
	model.Name = types.StringValue(api.Name)
	model.Uuid = types.StringValue(api.Uuid)
	model.Used = types.BoolValue(api.Used)
	model.Protocol = types.StringValue(api.Protocol)
	model.Comment = firewallCommentValue(api.Comment)
	model.TcpPortrange = listValue(ctx, types.StringType, model.TcpPortrange, api.TcpPortrange, diags)
	model.UdpPortrange = listValue(ctx, types.StringType, model.UdpPortrange, api.UdpPortrange, diags)
}
