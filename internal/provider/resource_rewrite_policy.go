package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/float64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RewritePolicyResource{}

type RewritePolicyResource struct {
	client *apiclient.Client
}

type RewritePolicyResourceModel struct {
	Id                     types.String  `tfsdk:"id"`
	Name                   types.String  `tfsdk:"name"`
	Rule                   types.String  `tfsdk:"rule"`
	Action                 types.String  `tfsdk:"action"`
	Undefaction            types.String  `tfsdk:"undefaction"`
	Comment                types.String  `tfsdk:"comment"`
	Priority               types.Float64 `tfsdk:"priority"`
	Bindpoint              types.String  `tfsdk:"bindpoint"`
	Gotopriorityexpression types.String  `tfsdk:"gotopriorityexpression"`
	Customer               types.String  `tfsdk:"customer"`
	Loadbalancer           types.String  `tfsdk:"loadbalancer"`
}

type rewritePolicyAPIModel struct {
	Id                     string   `json:"id,omitempty"`
	Name                   string   `json:"name"`
	Rule                   *string  `json:"rule,omitempty"`
	Action                 *string  `json:"action,omitempty"`
	Undefaction            *string  `json:"undefaction,omitempty"`
	Comment                *string  `json:"comment,omitempty"`
	Priority               *float64 `json:"priority,omitempty"`
	Bindpoint              *string  `json:"bindpoint,omitempty"`
	Gotopriorityexpression *string  `json:"gotopriorityexpression,omitempty"`
	Customer               *string  `json:"customer,omitempty"`
	Loadbalancer           *string  `json:"loadbalancer,omitempty"`
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
			"rule": schema.StringAttribute{
				Optional: true,
			},
			"action": schema.StringAttribute{
				Optional: true,
			},
			"undefaction": schema.StringAttribute{
				Optional: true,
			},
			"comment": schema.StringAttribute{
				Optional: true,
			},
			"priority": schema.Float64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Float64{
					float64planmodifier.UseStateForUnknown(),
				},
			},
			"bindpoint": schema.StringAttribute{
				Optional: true,
			},
			"gotopriorityexpression": schema.StringAttribute{
				Optional: true,
			},
			"customer": schema.StringAttribute{
				Optional: true,
			},
			"loadbalancer": schema.StringAttribute{
				Optional: true,
			},
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

func (r *RewritePolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RewritePolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := rewritePolicyAPIModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.Rule.IsNull() {
		v := plan.Rule.ValueString()
		apiModel.Rule = &v
	}
	if !plan.Action.IsNull() {
		v := plan.Action.ValueString()
		apiModel.Action = &v
	}
	if !plan.Undefaction.IsNull() {
		v := plan.Undefaction.ValueString()
		apiModel.Undefaction = &v
	}
	if !plan.Comment.IsNull() {
		v := plan.Comment.ValueString()
		apiModel.Comment = &v
	}
	if !plan.Priority.IsNull() {
		v := plan.Priority.ValueFloat64()
		apiModel.Priority = &v
	}
	if !plan.Bindpoint.IsNull() {
		v := plan.Bindpoint.ValueString()
		apiModel.Bindpoint = &v
	}
	if !plan.Gotopriorityexpression.IsNull() {
		v := plan.Gotopriorityexpression.ValueString()
		apiModel.Gotopriorityexpression = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp rewritePolicyAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/rewritepolicy/", apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating rewrite_policy", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Rule = types.StringPointerValue(apiResp.Rule)
	plan.Action = types.StringPointerValue(apiResp.Action)
	plan.Undefaction = types.StringPointerValue(apiResp.Undefaction)
	plan.Comment = types.StringPointerValue(apiResp.Comment)
	plan.Priority = types.Float64PointerValue(apiResp.Priority)
	plan.Bindpoint = types.StringPointerValue(apiResp.Bindpoint)
	plan.Gotopriorityexpression = types.StringPointerValue(apiResp.Gotopriorityexpression)
	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

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

	state.Id = types.StringValue(apiResp.Id)
	state.Name = types.StringValue(apiResp.Name)
	state.Rule = types.StringPointerValue(apiResp.Rule)
	state.Action = types.StringPointerValue(apiResp.Action)
	state.Undefaction = types.StringPointerValue(apiResp.Undefaction)
	state.Comment = types.StringPointerValue(apiResp.Comment)
	state.Priority = types.Float64PointerValue(apiResp.Priority)
	state.Bindpoint = types.StringPointerValue(apiResp.Bindpoint)
	state.Gotopriorityexpression = types.StringPointerValue(apiResp.Gotopriorityexpression)
	state.Customer = types.StringPointerValue(apiResp.Customer)
	state.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

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

	apiModel := rewritePolicyAPIModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.Rule.IsNull() {
		v := plan.Rule.ValueString()
		apiModel.Rule = &v
	}
	if !plan.Action.IsNull() {
		v := plan.Action.ValueString()
		apiModel.Action = &v
	}
	if !plan.Undefaction.IsNull() {
		v := plan.Undefaction.ValueString()
		apiModel.Undefaction = &v
	}
	if !plan.Comment.IsNull() {
		v := plan.Comment.ValueString()
		apiModel.Comment = &v
	}
	if !plan.Priority.IsNull() {
		v := plan.Priority.ValueFloat64()
		apiModel.Priority = &v
	}
	if !plan.Bindpoint.IsNull() {
		v := plan.Bindpoint.ValueString()
		apiModel.Bindpoint = &v
	}
	if !plan.Gotopriorityexpression.IsNull() {
		v := plan.Gotopriorityexpression.ValueString()
		apiModel.Gotopriorityexpression = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp rewritePolicyAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/rewritepolicy/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating rewrite_policy", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Rule = types.StringPointerValue(apiResp.Rule)
	plan.Action = types.StringPointerValue(apiResp.Action)
	plan.Undefaction = types.StringPointerValue(apiResp.Undefaction)
	plan.Comment = types.StringPointerValue(apiResp.Comment)
	plan.Priority = types.Float64PointerValue(apiResp.Priority)
	plan.Bindpoint = types.StringPointerValue(apiResp.Bindpoint)
	plan.Gotopriorityexpression = types.StringPointerValue(apiResp.Gotopriorityexpression)
	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

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
