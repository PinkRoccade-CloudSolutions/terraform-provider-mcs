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
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &RewriteActionResource{}
	_ resource.ResourceWithImportState = &RewriteActionResource{}
)

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

var rewriteActionTypes = []string{
	"replace", "replace_all", "replace_http_res", "insert_http_header", "delete_http_header",
	"corrupt_http_header", "insert_before", "insert_before_all", "insert_after", "insert_after_all",
	"delete", "delete_all", "clientless_vpn_encode", "clientless_vpn_encode_all",
	"clientless_vpn_decode", "clientless_vpn_decode_all", "replace_sip_res",
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
			"type": lbOptString("Rewrite action type (e.g. `replace`, `replace_all`, `insert_http_header`).",
				stringvalidator.OneOf(rewriteActionTypes...)),
			"target": lbOptString("Expression that specifies which part of the request or response to rewrite.",
				stringvalidator.LengthAtMost(16384)),
			"stringbuilderexpr": lbOptString("Expression that specifies the content to insert or replace.",
				stringvalidator.LengthAtMost(16384)),
			"search": lbOptString("Search pattern for _ALL action types.",
				stringvalidator.LengthAtMost(16384)),
			"comment":      lbOptString(""),
			"customer":     lbOptString(""),
			"loadbalancer": lbOptString(""),
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

func rewriteActionToAPI(plan *RewriteActionResourceModel) rewriteActionAPIModel {
	return rewriteActionAPIModel{
		Name:              plan.Name.ValueString(),
		Type:              stringPtr(plan.Type),
		Target:            stringPtr(plan.Target),
		Stringbuilderexpr: stringPtr(plan.Stringbuilderexpr),
		Search:            stringPtr(plan.Search),
		Comment:           stringPtr(plan.Comment),
		Customer:          stringPtr(plan.Customer),
		Loadbalancer:      stringPtr(plan.Loadbalancer),
	}
}

func rewriteActionFromAPI(m *RewriteActionResourceModel, api *rewriteActionAPIModel) {
	m.Id = types.StringValue(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Type = types.StringPointerValue(api.Type)
	m.Target = types.StringPointerValue(api.Target)
	m.Stringbuilderexpr = types.StringPointerValue(api.Stringbuilderexpr)
	m.Search = types.StringPointerValue(api.Search)
	m.Comment = types.StringPointerValue(api.Comment)
	m.Customer = types.StringPointerValue(api.Customer)
	m.Loadbalancer = types.StringPointerValue(api.Loadbalancer)
}

func (r *RewriteActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RewriteActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp rewriteActionAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/rewriteaction/", rewriteActionToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating rewrite_action", err.Error())
		return
	}

	rewriteActionFromAPI(&plan, &apiResp)
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

	rewriteActionFromAPI(&state, &apiResp)
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

	var apiResp rewriteActionAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/rewriteaction/%s/", state.Id.ValueString()), rewriteActionToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating rewrite_action", err.Error())
		return
	}

	rewriteActionFromAPI(&plan, &apiResp)
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

func (r *RewriteActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
