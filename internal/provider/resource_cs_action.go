package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ resource.Resource = &CsActionResource{}

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
	Lbvserver    *string `json:"lbvserver,omitempty"`
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

func (r *CsActionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CsActionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := csActionAPIModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.Lbvserver.IsNull() {
		v := plan.Lbvserver.ValueString()
		apiModel.Lbvserver = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp csActionAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/csaction/", apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating cs_action", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Lbvserver = types.StringPointerValue(apiResp.Lbvserver)
	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

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

	state.Id = types.StringValue(apiResp.Id)
	state.Name = types.StringValue(apiResp.Name)
	state.Lbvserver = types.StringPointerValue(apiResp.Lbvserver)
	state.Customer = types.StringPointerValue(apiResp.Customer)
	state.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

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

	apiModel := csActionAPIModel{
		Name: plan.Name.ValueString(),
	}
	if !plan.Lbvserver.IsNull() {
		v := plan.Lbvserver.ValueString()
		apiModel.Lbvserver = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp csActionAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/csaction/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating cs_action", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Lbvserver = types.StringPointerValue(apiResp.Lbvserver)
	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

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
