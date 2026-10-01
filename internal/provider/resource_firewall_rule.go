package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &FirewallRuleResource{}
	_ resource.ResourceWithImportState = &FirewallRuleResource{}
)

type FirewallRuleResource struct {
	client *apiclient.Client
}

type FirewallRuleModel struct {
	Id               types.String `tfsdk:"id"`
	Domain           types.String `tfsdk:"domain"`
	Enabled          types.Bool   `tfsdk:"enabled"`
	Src              types.List   `tfsdk:"src"`
	Dst              types.List   `tfsdk:"dst"`
	SrcIntf          types.List   `tfsdk:"src_intf"`
	DstIntf          types.List   `tfsdk:"dst_intf"`
	Service          types.List   `tfsdk:"service"`
	Action           types.Bool   `tfsdk:"action"`
	Uuid             types.String `tfsdk:"uuid"`
	PolicyId         types.Int64  `tfsdk:"policyid"`
	Group            types.String `tfsdk:"group"`
	Comment          types.String `tfsdk:"comment"`
	Origin           types.String `tfsdk:"origin"`
	Used             types.Bool   `tfsdk:"used"`
	Compliant        types.Bool   `tfsdk:"compliant"`
	HitCount         types.Int64  `tfsdk:"hit_count"`
	LastHit          types.String `tfsdk:"last_hit"`
	CompliancyErrors types.List   `tfsdk:"compliancy_errors"`
}

// firewallRuleRequest holds only the writable FirewallRule fields; policyid and the other
// read-only fields are never sent.
type firewallRuleRequest struct {
	Enabled bool      `json:"enabled"`
	Src     *[]string `json:"src,omitempty"`
	Dst     *[]string `json:"dst,omitempty"`
	SrcIntf *[]string `json:"src_intf,omitempty"`
	DstIntf *[]string `json:"dst_intf,omitempty"`
	Service *[]string `json:"service,omitempty"`
	Action  bool      `json:"action"`
	Comment *string   `json:"comment,omitempty"`
}

type firewallRuleAPI struct {
	Enabled          bool     `json:"enabled"`
	Src              []string `json:"src"`
	Dst              []string `json:"dst"`
	SrcIntf          []string `json:"src_intf"`
	DstIntf          []string `json:"dst_intf"`
	Service          []string `json:"service"`
	Action           bool     `json:"action"`
	Uuid             string   `json:"uuid"`
	PolicyId         int64    `json:"policyid"`
	Group            string   `json:"group"`
	Comment          *string  `json:"comment"`
	Origin           string   `json:"origin"`
	Used             bool     `json:"used"`
	Compliant        bool     `json:"compliant"`
	HitCount         int64    `json:"hit_count"`
	LastHit          string   `json:"last_hit"`
	CompliancyErrors []string `json:"compliancy_errors"`
}

func NewFirewallRuleResource() resource.Resource {
	return &FirewallRuleResource{}
}

func (r *FirewallRuleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall_rule"
}

func (r *FirewallRuleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a firewall rule in the MCS API.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"domain": schema.StringAttribute{
				Required:      true,
				Description:   "The domain this rule belongs to. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"enabled": schema.BoolAttribute{
				Required:    true,
				Description: "Whether the rule is enabled.",
			},
			"src": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Source addresses.",
			},
			"dst": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Destination addresses.",
			},
			"src_intf": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Source interfaces.",
			},
			"dst_intf": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Destination interfaces.",
			},
			"service": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Services for the rule.",
			},
			"action": schema.BoolAttribute{
				Required:    true,
				Description: "Action for the rule (true=allow, false=deny).",
			},
			"uuid": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID assigned by the firewall.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"policyid": schema.Int64Attribute{
				Computed:      true,
				Description:   "Policy ID assigned by the firewall.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"group": schema.StringAttribute{
				Computed:      true,
				Description:   "Group the rule belongs to.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"comment": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Comment for the rule.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"origin": schema.StringAttribute{
				Computed:      true,
				Description:   "Origin of the rule.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"used": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the rule is in use.",
			},
			"compliant": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the rule is compliant.",
			},
			"hit_count": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of hits on the rule.",
			},
			"last_hit": schema.StringAttribute{
				Computed:    true,
				Description: "Timestamp of the last hit on the rule.",
			},
			"compliancy_errors": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Compliancy errors reported for the rule.",
			},
		},
	}
}

func (r *FirewallRuleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *FirewallRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan FirewallRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := plan.Domain.ValueString()

	body := firewallRuleRequest{
		Enabled: plan.Enabled.ValueBool(),
		Action:  plan.Action.ValueBool(),
		Src:     listElems[string](ctx, plan.Src, &resp.Diagnostics),
		Dst:     listElems[string](ctx, plan.Dst, &resp.Diagnostics),
		SrcIntf: listElems[string](ctx, plan.SrcIntf, &resp.Diagnostics),
		DstIntf: listElems[string](ctx, plan.DstIntf, &resp.Diagnostics),
		Service: listElems[string](ctx, plan.Service, &resp.Diagnostics),
		Comment: stringPtr(plan.Comment),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallRuleAPI
	path := fmt.Sprintf("/api/networking/domain/%s/rules/", domain)
	if err := r.client.Post(ctx, path, body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating firewall rule", err.Error())
		return
	}

	mapFirewallRuleToState(ctx, &plan, domain, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state FirewallRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	policyid := strconv.FormatInt(state.PolicyId.ValueInt64(), 10)

	var apiResp firewallRuleAPI
	path := fmt.Sprintf("/api/networking/domain/%s/rules/%s/", domain, policyid)
	if err := r.client.Get(ctx, path, &apiResp); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading firewall rule", err.Error())
		return
	}

	mapFirewallRuleToState(ctx, &state, domain, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *FirewallRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state FirewallRuleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	policyid := strconv.FormatInt(state.PolicyId.ValueInt64(), 10)

	body := firewallRuleRequest{
		Enabled: plan.Enabled.ValueBool(),
		Action:  plan.Action.ValueBool(),
		Src:     listElemsForUpdate[string](ctx, plan.Src, state.Src, &resp.Diagnostics),
		Dst:     listElemsForUpdate[string](ctx, plan.Dst, state.Dst, &resp.Diagnostics),
		SrcIntf: listElemsForUpdate[string](ctx, plan.SrcIntf, state.SrcIntf, &resp.Diagnostics),
		DstIntf: listElemsForUpdate[string](ctx, plan.DstIntf, state.DstIntf, &resp.Diagnostics),
		Service: listElemsForUpdate[string](ctx, plan.Service, state.Service, &resp.Diagnostics),
		Comment: stringPtr(plan.Comment),
	}
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp firewallRuleAPI
	path := fmt.Sprintf("/api/networking/domain/%s/rules/%s/", domain, policyid)
	if err := r.client.Patch(ctx, path, body, &apiResp); err != nil {
		resp.Diagnostics.AddError("Error updating firewall rule", err.Error())
		return
	}

	mapFirewallRuleToState(ctx, &plan, domain, &apiResp, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *FirewallRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state FirewallRuleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain := state.Domain.ValueString()
	policyid := strconv.FormatInt(state.PolicyId.ValueInt64(), 10)

	path := fmt.Sprintf("/api/networking/domain/%s/rules/%s/", domain, policyid)
	if err := r.client.Delete(ctx, path); err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting firewall rule", err.Error())
	}
}

func mapFirewallRuleToState(ctx context.Context, model *FirewallRuleModel, domain string, api *firewallRuleAPI, diags *diag.Diagnostics) {
	model.Id = types.StringValue(strconv.FormatInt(api.PolicyId, 10))
	model.Domain = types.StringValue(domain)
	model.Enabled = types.BoolValue(api.Enabled)
	model.Action = types.BoolValue(api.Action)
	model.Uuid = types.StringValue(api.Uuid)
	model.PolicyId = types.Int64Value(api.PolicyId)
	model.Group = types.StringValue(api.Group)
	model.Comment = firewallCommentValue(api.Comment)
	model.Origin = types.StringValue(api.Origin)
	model.Used = types.BoolValue(api.Used)
	model.Compliant = types.BoolValue(api.Compliant)
	model.HitCount = types.Int64Value(api.HitCount)
	model.LastHit = types.StringValue(api.LastHit)
	model.CompliancyErrors = computedListValue(ctx, types.StringType, api.CompliancyErrors, diags)

	model.Src = listValue(ctx, types.StringType, model.Src, api.Src, diags)
	model.Dst = listValue(ctx, types.StringType, model.Dst, api.Dst, diags)
	model.SrcIntf = listValue(ctx, types.StringType, model.SrcIntf, api.SrcIntf, diags)
	model.DstIntf = listValue(ctx, types.StringType, model.DstIntf, api.DstIntf, diags)
	model.Service = listValue(ctx, types.StringType, model.Service, api.Service, diags)
}

// ImportState takes "<domain>/<policyid>", the two values the API addresses a rule by.
func (r *FirewallRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts, ok := parseImportID(req.ID, 2, "<domain>/<policyid>", &resp.Diagnostics)
	if !ok {
		return
	}
	policyid, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected an import ID of the form \"<domain>/<policyid>\" with a numeric policyid, got %q.", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("domain"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("policyid"), policyid)...)
}
