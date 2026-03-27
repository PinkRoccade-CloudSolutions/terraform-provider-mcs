package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &CsPolicyResource{}

type CsPolicyResource struct {
	client *apiclient.Client
}

type CsPolicyResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Action       types.String `tfsdk:"action"`
	Expression   types.String `tfsdk:"expression"`
	Customer     types.String `tfsdk:"customer"`
	Application  types.String `tfsdk:"application"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
}

type csPolicyAPIModel struct {
	Id           string  `json:"id,omitempty"`
	Name         string  `json:"name"`
	Action       *string `json:"action,omitempty"`
	Expression   *string `json:"expression,omitempty"`
	Customer     *string `json:"customer,omitempty"`
	Application  *string `json:"application,omitempty"`
	Loadbalancer *string `json:"loadbalancer,omitempty"`
}

func NewCsPolicyResource() resource.Resource {
	return &CsPolicyResource{}
}

func (r *CsPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cs_policy"
}

func (r *CsPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"action": schema.StringAttribute{
				Optional: true,
			},
			"expression": schema.StringAttribute{
				Optional: true,
			},
			"customer": schema.StringAttribute{
				Optional: true,
			},
			"application": schema.StringAttribute{
				Optional: true,
			},
			"loadbalancer": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *CsPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CsPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CsPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := csPolicyAPIModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.Action.IsNull() {
		v := plan.Action.ValueString()
		apiModel.Action = &v
	}
	if !plan.Expression.IsNull() {
		v := plan.Expression.ValueString()
		apiModel.Expression = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Application.IsNull() {
		v := plan.Application.ValueString()
		apiModel.Application = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp csPolicyAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/cspolicy/", apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating cs_policy", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Action = types.StringPointerValue(apiResp.Action)
	plan.Expression = types.StringPointerValue(apiResp.Expression)
	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Application = types.StringPointerValue(apiResp.Application)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CsPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CsPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp csPolicyAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/cspolicy/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading cs_policy", err.Error())
		return
	}

	state.Id = types.StringValue(apiResp.Id)
	state.Name = types.StringValue(apiResp.Name)
	state.Action = types.StringPointerValue(apiResp.Action)
	state.Expression = types.StringPointerValue(apiResp.Expression)
	state.Customer = types.StringPointerValue(apiResp.Customer)
	state.Application = types.StringPointerValue(apiResp.Application)
	state.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CsPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CsPolicyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CsPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := csPolicyAPIModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.Action.IsNull() {
		v := plan.Action.ValueString()
		apiModel.Action = &v
	}
	if !plan.Expression.IsNull() {
		v := plan.Expression.ValueString()
		apiModel.Expression = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Application.IsNull() {
		v := plan.Application.ValueString()
		apiModel.Application = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp csPolicyAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/cspolicy/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating cs_policy", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Action = types.StringPointerValue(apiResp.Action)
	plan.Expression = types.StringPointerValue(apiResp.Expression)
	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Application = types.StringPointerValue(apiResp.Application)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CsPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CsPolicyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/cspolicy/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting cs_policy", err.Error())
	}
}
