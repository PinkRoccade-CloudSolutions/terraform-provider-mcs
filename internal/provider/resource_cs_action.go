package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &CsActionResource{}
	_ resource.ResourceWithImportState = &CsActionResource{}
)

type CsActionResource struct {
	client *apiclient.Client
}

type CsActionResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Lbvserver    types.String `tfsdk:"lbvserver"`
	Customer     types.String `tfsdk:"customer"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
}

type csActionAPIModel struct {
	Id           string  `json:"id,omitempty"`
	Name         string  `json:"name"`
	Lbvserver    *string `json:"lbvserver"`
	Customer     *string `json:"customer,omitempty"`
	Loadbalancer *string `json:"loadbalancer,omitempty"`
}

func NewCsActionResource() resource.Resource {
	return &CsActionResource{}
}

func (r *CsActionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cs_action"
}

func (r *CsActionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"lbvserver": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the target LB vServer.",
			},
			"customer":     lbOptString(""),
			"loadbalancer": lbOptString(""),
		},
	}
}

func (r *CsActionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func csActionToAPI(plan *CsActionResourceModel) csActionAPIModel {
	return csActionAPIModel{
		Name:         plan.Name.ValueString(),
		Lbvserver:    stringPtr(plan.Lbvserver),
		Customer:     stringPtr(plan.Customer),
		Loadbalancer: stringPtr(plan.Loadbalancer),
	}
}

func csActionFromAPI(m *CsActionResourceModel, api *csActionAPIModel) {
	m.Id = types.StringValue(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Lbvserver = types.StringPointerValue(api.Lbvserver)
	m.Customer = types.StringPointerValue(api.Customer)
	m.Loadbalancer = types.StringPointerValue(api.Loadbalancer)
}

func (r *CsActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CsActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp csActionAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/csaction/", csActionToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating cs_action", err.Error())
		return
	}

	csActionFromAPI(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CsActionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CsActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp csActionAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/csaction/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading cs_action", err.Error())
		return
	}

	csActionFromAPI(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CsActionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CsActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CsActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp csActionAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/csaction/%s/", state.Id.ValueString()), csActionToAPI(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating cs_action", err.Error())
		return
	}

	csActionFromAPI(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CsActionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CsActionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/csaction/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting cs_action", err.Error())
	}
}

func (r *CsActionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
