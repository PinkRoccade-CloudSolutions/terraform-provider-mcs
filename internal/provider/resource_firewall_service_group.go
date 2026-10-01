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

var _ resource.Resource = &FirewallServiceGroupResource{}

type FirewallServiceGroupResource struct {
	client *apiclient.Client
}

type FirewallServiceGroupModel struct {
	Id      types.String `tfsdk:"id"`
	Domain  types.String `tfsdk:"domain"`
	Name    types.String `tfsdk:"name"`
	Uuid    types.String `tfsdk:"uuid"`
	Comment types.String `tfsdk:"comment"`
	Member  types.List   `tfsdk:"member"`
	Used    types.Bool   `tfsdk:"used"`
}

type firewallServiceGroupRequest struct {
	Name    string    `json:"name"`
	Comment *string   `json:"comment,omitempty"`
	Member  *[]string `json:"member,omitempty"`
}

type firewallServiceGroupAPI struct {
	Name    string   `json:"name"`
	Uuid    string   `json:"uuid"`
	Comment *string  `json:"comment"`
	Member  []string `json:"member"`
	Used    bool     `json:"used"`
}

func NewFirewallServiceGroupResource() resource.Resource {
	return &FirewallServiceGroupResource{}
}

func (r *FirewallServiceGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_service_group"
}

func (r *FirewallServiceGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a firewall service group in the MCS API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain": schema.StringAttribute{
				Required:      true,
				Description:   "The domain this service group belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the service group.",
			},
			"uuid": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID assigned by the firewall.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"comment": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Comment for the service group.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"member": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of member service names. Removing the attribute clears the members.",
			},
			"used": schema.BoolAttribute{
				Computed:      true,
				Description:   "Whether the service group is currently in use.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *FirewallServiceGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func mapFirewallServiceGroupToState(ctx context.Context, model *FirewallServiceGroupModel, api *firewallServiceGroupAPI, diags *diag.Diagnostics) {
	model.Id = types.StringValue(api.Uuid)
	model.Name = types.StringValue(api.Name)
	model.Uuid = types.StringValue(api.Uuid)
	model.Used = types.BoolValue(api.Used)
	model.Comment = firewallCommentValue(api.Comment)
	model.Member = listValue(ctx, types.StringType, model.Member, api.Member, diags)
}

func (r *FirewallServiceGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallServiceGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := firewallServiceGroupRequest{
		Name:    plan.Name.ValueString(),
		Comment: stringPtr(plan.Comment),
		Member:  listElems[string](ctx, plan.Member, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallServiceGroupAPI
	path := fmt.Sprintf("/api/networking/domain/%s/servicegroups/", plan.Domain.ValueString())
	if err := r.client.Post(ctx, path, body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating firewall service group", err.Error())
		return
	}

	mapFirewallServiceGroupToState(ctx, &plan, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallServiceGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallServiceGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallServiceGroupAPI
	path := fmt.Sprintf("/api/networking/domain/%s/servicegroups/%s/", state.Domain.ValueString(), state.Name.ValueString())
	if err := r.client.Get(ctx, path, &apiResp); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall service group", err.Error())
		return
	}

	mapFirewallServiceGroupToState(ctx, &state, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallServiceGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state FirewallServiceGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := firewallServiceGroupRequest{
		Name:    plan.Name.ValueString(),
		Comment: stringPtr(plan.Comment),
		Member:  listElemsForUpdate[string](ctx, plan.Member, state.Member, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Address the group by its current (state) name so that renames work.
	var apiResp firewallServiceGroupAPI
	path := fmt.Sprintf("/api/networking/domain/%s/servicegroups/%s/", state.Domain.ValueString(), state.Name.ValueString())
	if err := r.client.Patch(ctx, path, body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error updating firewall service group", err.Error())
		return
	}

	mapFirewallServiceGroupToState(ctx, &plan, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallServiceGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallServiceGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("/api/networking/domain/%s/servicegroups/%s/", state.Domain.ValueString(), state.Name.ValueString())
	if err := r.client.Delete(ctx, path); err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting firewall service group", err.Error())
	}
}
