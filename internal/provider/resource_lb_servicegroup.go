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

var _ resource.Resource = &LbServicegroupResource{}

type LbServicegroupResource struct {
	client *apiclient.Client
}

type LbServicegroupResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Type          types.String `tfsdk:"type"`
	State         types.String `tfsdk:"state"`
	Members       types.List   `tfsdk:"members"`
	Healthmonitor types.String `tfsdk:"healthmonitor"`
	Customer      types.String `tfsdk:"customer"`
	Loadbalancer  types.String `tfsdk:"loadbalancer"`
}

type lbServicegroupAPIModel struct {
	Id            string   `json:"id,omitempty"`
	Name          string   `json:"name"`
	Type          string   `json:"type"`
	State         *string  `json:"state,omitempty"`
	Members       []string `json:"members,omitempty"`
	Healthmonitor *string  `json:"healthmonitor,omitempty"`
	Customer      *string  `json:"customer,omitempty"`
	Loadbalancer  *string  `json:"loadbalancer,omitempty"`
}

func NewLbServicegroupResource() resource.Resource {
	return &LbServicegroupResource{}
}

func (r *LbServicegroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lb_servicegroup"
}

func (r *LbServicegroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Required: true,
			},
			"state": schema.StringAttribute{
				Optional: true,
			},
			"members": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"healthmonitor": schema.StringAttribute{
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

func (r *LbServicegroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LbServicegroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LbServicegroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := lbServicegroupAPIModel{
		Name: plan.Name.ValueString(),
		Type: plan.Type.ValueString(),
	}
	if !plan.State.IsNull() {
		v := plan.State.ValueString()
		apiModel.State = &v
	}
	if !plan.Members.IsNull() {
		var members []string
		resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &members, false)...)
		apiModel.Members = members
	}
	if !plan.Healthmonitor.IsNull() {
		v := plan.Healthmonitor.ValueString()
		apiModel.Healthmonitor = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp lbServicegroupAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/lbservicegroup/", apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating lb_servicegroup", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Type = types.StringValue(apiResp.Type)
	plan.State = types.StringPointerValue(apiResp.State)

	listVal, diags := types.ListValueFrom(ctx, types.StringType, apiResp.Members)
	resp.Diagnostics.Append(diags...)
	plan.Members = listVal

	plan.Healthmonitor = types.StringPointerValue(apiResp.Healthmonitor)
	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LbServicegroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LbServicegroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbServicegroupAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroup/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading lb_servicegroup", err.Error())
		return
	}

	state.Id = types.StringValue(apiResp.Id)
	state.Name = types.StringValue(apiResp.Name)
	state.Type = types.StringValue(apiResp.Type)
	state.State = types.StringPointerValue(apiResp.State)

	listVal, diags := types.ListValueFrom(ctx, types.StringType, apiResp.Members)
	resp.Diagnostics.Append(diags...)
	state.Members = listVal

	state.Healthmonitor = types.StringPointerValue(apiResp.Healthmonitor)
	state.Customer = types.StringPointerValue(apiResp.Customer)
	state.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LbServicegroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LbServicegroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state LbServicegroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := lbServicegroupAPIModel{
		Name: plan.Name.ValueString(),
		Type: plan.Type.ValueString(),
	}
	if !plan.State.IsNull() {
		v := plan.State.ValueString()
		apiModel.State = &v
	}
	if !plan.Members.IsNull() {
		var members []string
		resp.Diagnostics.Append(plan.Members.ElementsAs(ctx, &members, false)...)
		apiModel.Members = members
	}
	if !plan.Healthmonitor.IsNull() {
		v := plan.Healthmonitor.ValueString()
		apiModel.Healthmonitor = &v
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp lbServicegroupAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroup/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating lb_servicegroup", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Type = types.StringValue(apiResp.Type)
	plan.State = types.StringPointerValue(apiResp.State)

	listVal, diags := types.ListValueFrom(ctx, types.StringType, apiResp.Members)
	resp.Diagnostics.Append(diags...)
	plan.Members = listVal

	plan.Healthmonitor = types.StringPointerValue(apiResp.Healthmonitor)
	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LbServicegroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LbServicegroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/lbservicegroup/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting lb_servicegroup", err.Error())
	}
}
