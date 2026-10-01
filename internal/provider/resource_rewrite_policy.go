package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RewritePolicyResource{}

type RewritePolicyResource struct {
	client *apiclient.Client
}

type RewritePolicyResourceModel struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Rule                   types.String `tfsdk:"rule"`
	Action                 types.String `tfsdk:"action"`
	Undefaction            types.String `tfsdk:"undefaction"`
	Comment                types.String `tfsdk:"comment"`
	Priority               types.Int64  `tfsdk:"priority"`
	Bindpoint              types.String `tfsdk:"bindpoint"`
	Gotopriorityexpression types.String `tfsdk:"gotopriorityexpression"`
	Customer               types.String `tfsdk:"customer"`
	Loadbalancer           types.String `tfsdk:"loadbalancer"`
}

type rewritePolicyAPIModel struct {
	Id                     string  `json:"id,omitempty"`
	Name                   string  `json:"name"`
	Rule                   *string `json:"rule,omitempty"`
	Action                 *string `json:"action"`
	Undefaction            *string `json:"undefaction,omitempty"`
	Comment                *string `json:"comment,omitempty"`
	Priority               *int64  `json:"priority,omitempty"`
	Bindpoint              *string `json:"bindpoint,omitempty"`
	Gotopriorityexpression *string `json:"gotopriorityexpression,omitempty"`
	Customer               *string `json:"customer,omitempty"`
	Loadbalancer           *string `json:"loadbalancer,omitempty"`
}

func NewRewritePolicyResource() resource.Resource {
	return &RewritePolicyResource{}
}

func (r *RewritePolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rewrite_policy"
}

func (r *RewritePolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"rule": lbOptString("Expression to evaluate (e.g. HTTP.REQ.URL.CONTAINS(\"test\")).",
				stringvalidator.LengthAtMost(16384)),
			"action": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the rewrite action to perform when the rule matches.",
			},
			"undefaction": lbOptString("Action if the result is undefined (NOREWRITE, RESET, DROP).",
				stringvalidator.LengthAtMost(255)),
			"comment":  lbOptString(""),
			"priority": lbOptInt64("Priority of the policy binding. Lower number = higher priority."),
			"bindpoint": lbOptString("Bind to `REQUEST` or `RESPONSE` processing.",
				stringvalidator.OneOf("REQUEST", "RESPONSE")),
			"gotopriorityexpression": lbOptString("Action after this policy: `NEXT`, `END` or `USE_INVOCATION_RESULT`.",
				stringvalidator.OneOf("NEXT", "END", "USE_INVOCATION_RESULT")),
			"customer":     lbOptString(""),
			"loadbalancer": lbOptString(""),
		},
	}
}

func (r *RewritePolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T", req.ProviderData),
		)
		return
	}
	r.client = client
}

func rewritePolicyToAPI(plan *RewritePolicyResourceModel) rewritePolicyAPIModel {
	return rewritePolicyAPIModel{
		Name:                   plan.Name.ValueString(),
		Rule:                   stringPtr(plan.Rule),
		Action:                 stringPtr(plan.Action),
		Undefaction:            stringPtr(plan.Undefaction),
		Comment:                stringPtr(plan.Comment),
		Priority:               int64Ptr(plan.Priority),
		Bindpoint:              stringPtr(plan.Bindpoint),
		Gotopriorityexpression: stringPtr(plan.Gotopriorityexpression),
		Customer:               stringPtr(plan.Customer),
		Loadbalancer:           stringPtr(plan.Loadbalancer),
	}
}

func rewritePolicyFromAPI(m *RewritePolicyResourceModel, api *rewritePolicyAPIModel) {
	m.Id = types.StringValue(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Rule = types.StringPointerValue(api.Rule)
	m.Action = types.StringPointerValue(api.Action)
	m.Undefaction = types.StringPointerValue(api.Undefaction)
	m.Comment = types.StringPointerValue(api.Comment)
	m.Priority = types.Int64PointerValue(api.Priority)
	m.Bindpoint = types.StringPointerValue(api.Bindpoint)
	m.Gotopriorityexpression = types.StringPointerValue(api.Gotopriorityexpression)
	m.Customer = types.StringPointerValue(api.Customer)
	m.Loadbalancer = types.StringPointerValue(api.Loadbalancer)
}

func (r *RewritePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RewritePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp rewritePolicyAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/rewritepolicy/", rewritePolicyToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating rewrite_policy", err.Error())
		return
	}

	rewritePolicyFromAPI(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RewritePolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RewritePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp rewritePolicyAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/rewritepolicy/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading rewrite_policy", err.Error())
		return
	}

	rewritePolicyFromAPI(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RewritePolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RewritePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state RewritePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp rewritePolicyAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/rewritepolicy/%s/", state.Id.ValueString()), rewritePolicyToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating rewrite_policy", err.Error())
		return
	}

	rewritePolicyFromAPI(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RewritePolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RewritePolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/rewritepolicy/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting rewrite_policy", err.Error())
	}
}
