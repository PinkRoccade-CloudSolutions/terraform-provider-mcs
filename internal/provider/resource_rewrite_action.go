package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &RewriteActionResource{}

type RewriteActionResource struct {
	client *apiclient.Client
}

type RewriteActionResourceModel struct {
	Id                types.String `tfsdk:"id"`
	Name              types.String `tfsdk:"name"`
	Type              types.String `tfsdk:"type"`
	Target            types.String `tfsdk:"target"`
	Stringbuilderexpr types.String `tfsdk:"stringbuilderexpr"`
	Search            types.String `tfsdk:"search"`
	Comment           types.String `tfsdk:"comment"`
	Customer          types.String `tfsdk:"customer"`
	Loadbalancer      types.String `tfsdk:"loadbalancer"`
}

type rewriteActionAPIModel struct {
	Id                string  `json:"id,omitempty"`
	Name              string  `json:"name"`
	Type              *string `json:"type,omitempty"`
	Target            *string `json:"target,omitempty"`
	Stringbuilderexpr *string `json:"stringbuilderexpr,omitempty"`
	Search            *string `json:"search,omitempty"`
	Comment           *string `json:"comment,omitempty"`
	Customer          *string `json:"customer,omitempty"`
	Loadbalancer      *string `json:"loadbalancer,omitempty"`
}

func NewRewriteActionResource() resource.Resource {
	return &RewriteActionResource{}
}

func (r *RewriteActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_rewrite_action"
}

func (r *RewriteActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Optional: true,
			},
			"target": schema.StringAttribute{
				Optional: true,
			},
			"stringbuilderexpr": schema.StringAttribute{
				Optional: true,
			},
			"search": schema.StringAttribute{
				Optional: true,
			},
			"comment": schema.StringAttribute{
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

func (r *RewriteActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RewriteActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RewriteActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := rewriteActionAPIModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.Type.IsNull() {
		v := plan.Type.ValueString()
		apiModel.Type = &v
	}
	if !plan.Target.IsNull() {
		v := plan.Target.ValueString()
		apiModel.Target = &v
	}
	if !plan.Stringbuilderexpr.IsNull() {
		v := plan.Stringbuilderexpr.ValueString()
		apiModel.Stringbuilderexpr = &v
	}
	if !plan.Search.IsNull() {
		v := plan.Search.ValueString()
		apiModel.Search = &v
	}
	if !plan.Comment.IsNull() {
		v := plan.Comment.ValueString()
		apiModel.Comment = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp rewriteActionAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/rewriteaction/", apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating rewrite_action", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Type = types.StringPointerValue(apiResp.Type)
	plan.Target = types.StringPointerValue(apiResp.Target)
	plan.Stringbuilderexpr = types.StringPointerValue(apiResp.Stringbuilderexpr)
	plan.Search = types.StringPointerValue(apiResp.Search)
	plan.Comment = types.StringPointerValue(apiResp.Comment)
	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RewriteActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RewriteActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp rewriteActionAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/rewriteaction/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading rewrite_action", err.Error())
		return
	}

	state.Id = types.StringValue(apiResp.Id)
	state.Name = types.StringValue(apiResp.Name)
	state.Type = types.StringPointerValue(apiResp.Type)
	state.Target = types.StringPointerValue(apiResp.Target)
	state.Stringbuilderexpr = types.StringPointerValue(apiResp.Stringbuilderexpr)
	state.Search = types.StringPointerValue(apiResp.Search)
	state.Comment = types.StringPointerValue(apiResp.Comment)
	state.Customer = types.StringPointerValue(apiResp.Customer)
	state.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *RewriteActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RewriteActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state RewriteActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := rewriteActionAPIModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.Type.IsNull() {
		v := plan.Type.ValueString()
		apiModel.Type = &v
	}
	if !plan.Target.IsNull() {
		v := plan.Target.ValueString()
		apiModel.Target = &v
	}
	if !plan.Stringbuilderexpr.IsNull() {
		v := plan.Stringbuilderexpr.ValueString()
		apiModel.Stringbuilderexpr = &v
	}
	if !plan.Search.IsNull() {
		v := plan.Search.ValueString()
		apiModel.Search = &v
	}
	if !plan.Comment.IsNull() {
		v := plan.Comment.ValueString()
		apiModel.Comment = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp rewriteActionAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/rewriteaction/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating rewrite_action", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Type = types.StringPointerValue(apiResp.Type)
	plan.Target = types.StringPointerValue(apiResp.Target)
	plan.Stringbuilderexpr = types.StringPointerValue(apiResp.Stringbuilderexpr)
	plan.Search = types.StringPointerValue(apiResp.Search)
	plan.Comment = types.StringPointerValue(apiResp.Comment)
	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *RewriteActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RewriteActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/rewriteaction/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting rewrite_action", err.Error())
	}
}
