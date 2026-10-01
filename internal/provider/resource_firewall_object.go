package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &FirewallObjectResource{}
	_ resource.ResourceWithImportState = &FirewallObjectResource{}
)

type FirewallObjectResource struct {
	client *apiclient.Client
}

type FirewallObjectModel struct {
	Id      types.String `tfsdk:"id"`
	Domain  types.String `tfsdk:"domain"`
	Name    types.String `tfsdk:"name"`
	Uuid    types.String `tfsdk:"uuid"`
	Address types.String `tfsdk:"address"`
	Subnet  types.String `tfsdk:"subnet"`
	Comment types.String `tfsdk:"comment"`
	Used    types.Bool   `tfsdk:"used"`
	Managed types.Bool   `tfsdk:"managed"`
}

type firewallObjectRequest struct {
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Subnet  string  `json:"subnet"`
	Comment *string `json:"comment,omitempty"`
}

type firewallObjectAPI struct {
	Name    string  `json:"name"`
	Uuid    string  `json:"uuid"`
	Address string  `json:"address"`
	Subnet  string  `json:"subnet"`
	Comment *string `json:"comment"`
	Used    bool    `json:"used"`
	Managed bool    `json:"managed"`
}

func NewFirewallObjectResource() resource.Resource {
	return &FirewallObjectResource{}
}

func (r *FirewallObjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_object"
}

func (r *FirewallObjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a firewall address object in the MCS API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain": schema.StringAttribute{
				Required:      true,
				Description:   "The domain this object belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the firewall object.",
			},
			"uuid": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID assigned by the firewall.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"address": schema.StringAttribute{
				Required:    true,
				Description: "IP address of the object.",
			},
			"subnet": schema.StringAttribute{
				Required:    true,
				Description: "Subnet mask of the object.",
			},
			"comment": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Comment for the object.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"used": schema.BoolAttribute{
				Computed:      true,
				Description:   "Whether the object is currently in use.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"managed": schema.BoolAttribute{
				Computed:      true,
				Description:   "Whether the object is managed by MCS.",
				PlanModifiers: []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *FirewallObjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func firewallObjectBody(plan *FirewallObjectModel) firewallObjectRequest {
	return firewallObjectRequest{
		Name:    plan.Name.ValueString(),
		Address: plan.Address.ValueString(),
		Subnet:  plan.Subnet.ValueString(),
		Comment: stringPtr(plan.Comment),
	}
}

func mapFirewallObjectToState(model *FirewallObjectModel, api *firewallObjectAPI) {
	model.Id = types.StringValue(api.Uuid)
	model.Name = types.StringValue(api.Name)
	model.Uuid = types.StringValue(api.Uuid)
	model.Used = types.BoolValue(api.Used)
	model.Managed = types.BoolValue(api.Managed)
	model.Address = types.StringValue(api.Address)
	model.Subnet = types.StringValue(api.Subnet)
	model.Comment = firewallCommentValue(api.Comment)
}

// firewallCommentValue maps an Optional+Computed comment; an absent comment is stored as "".
func firewallCommentValue(v *string) types.String {
	if v == nil {
		return types.StringValue("")
	}
	return types.StringValue(*v)
}

func (r *FirewallObjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallObjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallObjectAPI
	path := fmt.Sprintf("/api/networking/domain/%s/objects/", plan.Domain.ValueString())
	if err := r.client.Post(ctx, path, firewallObjectBody(&plan), &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating firewall object", err.Error())
		return
	}

	mapFirewallObjectToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallObjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallObjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallObjectAPI
	path := fmt.Sprintf("/api/networking/domain/%s/objects/%s/", state.Domain.ValueString(), state.Name.ValueString())
	if err := r.client.Get(ctx, path, &apiResp); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall object", err.Error())
		return
	}

	mapFirewallObjectToState(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallObjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state FirewallObjectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Address the object by its current (state) name so that renames work.
	var apiResp firewallObjectAPI
	path := fmt.Sprintf("/api/networking/domain/%s/objects/%s/", state.Domain.ValueString(), state.Name.ValueString())
	if err := r.client.Patch(ctx, path, firewallObjectBody(&plan), &apiResp); err != nil {
		resp.Diagnostics.AddError("Error updating firewall object", err.Error())
		return
	}

	mapFirewallObjectToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallObjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallObjectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := fmt.Sprintf("/api/networking/domain/%s/objects/%s/", state.Domain.ValueString(), state.Name.ValueString())
	if err := r.client.Delete(ctx, path); err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting firewall object", err.Error())
	}
}

// ImportState takes "<domain>/<name>", the two values the API addresses this object by.
func (r *FirewallObjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, ok := parseImportID(req.ID, 2, "<domain>/<name>", &resp.Diagnostics)
	if !ok {
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), parts[1])...)
}
