package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &FirewallObjectGroupResource{}
	_ resource.ResourceWithImportState = &FirewallObjectGroupResource{}
)

type FirewallObjectGroupResource struct {
	client *apiclient.Client
}

type FirewallObjectGroupModel struct {
	Id      types.String `tfsdk:"id"`
	Domain  types.String `tfsdk:"domain"`
	Name    types.String `tfsdk:"name"`
	Uuid    types.String `tfsdk:"uuid"`
	Comment types.String `tfsdk:"comment"`
	Member  types.List   `tfsdk:"member"`
	Used    types.Bool   `tfsdk:"used"`
}

type firewallObjectGroupRequest struct {
	Name    string    `json:"name"`
	Comment *string   `json:"comment,omitempty"`
	Member  *[]string `json:"member,omitempty"`
}

type firewallObjectGroupAPI struct {
	Name    string   `json:"name"`
	Uuid    string   `json:"uuid"`
	Comment *string  `json:"comment"`
	Member  []string `json:"member"`
	Used    bool     `json:"used"`
}

func NewFirewallObjectGroupResource() resource.Resource {
	return &FirewallObjectGroupResource{}
}

func (r *FirewallObjectGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_object_group"
}

func (r *FirewallObjectGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a firewall object group in the MCS API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain": schema.StringAttribute{
				Required:      true,
				Description:   "The domain this group belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the object group.",
			},
			"uuid": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID assigned by the firewall.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"comment": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Comment for the group.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"member": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of member object names. Removing the attribute clears the members.",
			},
			"used": schema.BoolAttribute{
				Computed:      true,
				Description:   "Whether the group is currently in use.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *FirewallObjectGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func mapFirewallObjectGroupToState(ctx context.Context, model *FirewallObjectGroupModel, api *firewallObjectGroupAPI, diags *diag.Diagnostics) {
	model.Id = types.StringValue(api.Uuid)
	model.Name = types.StringValue(api.Name)
	model.Uuid = types.StringValue(api.Uuid)
	model.Used = types.BoolValue(api.Used)
	model.Comment = firewallCommentValue(api.Comment)
	model.Member = listValue(ctx, types.StringType, model.Member, api.Member, diags)
}

func (r *FirewallObjectGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallObjectGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := firewallObjectGroupRequest{
		Name:    plan.Name.ValueString(),
		Comment: stringPtr(plan.Comment),
		Member:  listElems[string](ctx, plan.Member, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallObjectGroupAPI
	path := fmt.Sprintf("/api/networking/domain/%s/groups/", plan.Domain.ValueString())
	if err := r.client.Post(ctx, path, body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating firewall object group", err.Error())
		return
	}

	mapFirewallObjectGroupToState(ctx, &plan, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallObjectGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallObjectGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallObjectGroupAPI
	path := fmt.Sprintf("/api/networking/domain/%s/groups/%s/", state.Domain.ValueString(), state.Name.ValueString())
	if err := r.client.Get(ctx, path, &apiResp); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall object group", err.Error())
		return
	}

	mapFirewallObjectGroupToState(ctx, &state, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallObjectGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state FirewallObjectGroupModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	body := firewallObjectGroupRequest{
		Name:    plan.Name.ValueString(),
		Comment: stringPtr(plan.Comment),
		Member:  listElemsForUpdate[string](ctx, plan.Member, state.Member, &resp.Diagnostics),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	// Address the group by its current (state) name so that renames work.
	var apiResp firewallObjectGroupAPI
	path := fmt.Sprintf("/api/networking/domain/%s/groups/%s/", state.Domain.ValueString(), state.Name.ValueString())
	if err := r.client.Patch(ctx, path, body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error updating firewall object group", err.Error())
		return
	}

	mapFirewallObjectGroupToState(ctx, &plan, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallObjectGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallObjectGroupModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("/api/networking/domain/%s/groups/%s/", state.Domain.ValueString(), state.Name.ValueString())
	if err := r.client.Delete(ctx, path); err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting firewall object group", err.Error())
	}
}

// ImportState takes "<domain>/<name>", the two values the API addresses this object by.
func (r *FirewallObjectGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, ok := parseImportID(req.ID, 2, "<domain>/<name>", &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}
