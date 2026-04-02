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

var _ resource.Resource = &LbvServerResource{}

type LbvServerResource struct {
	client *apiclient.Client
}

type LbvServerResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Ipaddress    types.String `tfsdk:"ipaddress"`
	Port         types.Int64  `tfsdk:"port"`
	Type         types.String `tfsdk:"type"`
	Servicegroup types.List   `tfsdk:"servicegroup"`
	Certificate  types.List   `tfsdk:"certificate"`
	Customer     types.String `tfsdk:"customer"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
}

type lbvServerAPIModel struct {
	Id           string   `json:"id,omitempty"`
	Name         string   `json:"name"`
	Ipaddress    *string  `json:"ipaddress,omitempty"`
	Port         *int64   `json:"port,omitempty"`
	Type         *string  `json:"type,omitempty"`
	Servicegroup []string `json:"servicegroup"`
	Certificate  []string `json:"certificate,omitempty"`
	Customer     *string  `json:"customer,omitempty"`
	Loadbalancer *string  `json:"loadbalancer,omitempty"`
}

func NewLbvServerResource() resource.Resource {
	return &LbvServerResource{}
}

func (r *LbvServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lbv_server"
}

func (r *LbvServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"ipaddress": schema.StringAttribute{
				Optional:    true,
				Description: "UUID of the associated PublicIPAddress. Leave empty if used as non routed loadbalancer.",
			},
			"port": schema.Int64Attribute{
				Optional: true,
			},
			"type": schema.StringAttribute{
				Optional: true,
			},
			"servicegroup": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"certificate": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
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

func (r *LbvServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *LbvServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LbvServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := lbvServerAPIModel{
		Name: plan.Name.ValueString(),
	}

	var servicegroup []string
	resp.Diagnostics.Append(plan.Servicegroup.ElementsAs(ctx, &servicegroup, false)...)
	apiModel.Servicegroup = servicegroup

	if !plan.Ipaddress.IsNull() {
		v := plan.Ipaddress.ValueString()
		apiModel.Ipaddress = &v
	}
	if !plan.Port.IsNull() {
		v := plan.Port.ValueInt64()
		apiModel.Port = &v
	}
	if !plan.Type.IsNull() {
		v := plan.Type.ValueString()
		apiModel.Type = &v
	}
	if !plan.Certificate.IsNull() {
		var certificate []string
		resp.Diagnostics.Append(plan.Certificate.ElementsAs(ctx, &certificate, false)...)
		apiModel.Certificate = certificate
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp lbvServerAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/lbvserver/", apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating lbv_server", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	if apiResp.Ipaddress != nil {
		plan.Ipaddress = types.StringPointerValue(apiResp.Ipaddress)
	}
	plan.Port = types.Int64PointerValue(apiResp.Port)
	plan.Type = types.StringPointerValue(apiResp.Type)

	listVal, diags := types.ListValueFrom(ctx, types.StringType, apiResp.Servicegroup)
	resp.Diagnostics.Append(diags...)
	plan.Servicegroup = listVal

	if plan.Certificate.IsNull() && len(apiResp.Certificate) == 0 {
		plan.Certificate = types.ListNull(types.StringType)
	} else {
		listVal, diags = types.ListValueFrom(ctx, types.StringType, apiResp.Certificate)
		resp.Diagnostics.Append(diags...)
		plan.Certificate = listVal
	}

	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LbvServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state LbvServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbvServerAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/lbvserver/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading lbv_server", err.Error())
		return
	}

	state.Id = types.StringValue(apiResp.Id)
	state.Name = types.StringValue(apiResp.Name)
	if apiResp.Ipaddress != nil {
		state.Ipaddress = types.StringPointerValue(apiResp.Ipaddress)
	}
	state.Port = types.Int64PointerValue(apiResp.Port)
	state.Type = types.StringPointerValue(apiResp.Type)

	listVal, diags := types.ListValueFrom(ctx, types.StringType, apiResp.Servicegroup)
	resp.Diagnostics.Append(diags...)
	state.Servicegroup = listVal

	if state.Certificate.IsNull() && len(apiResp.Certificate) == 0 {
		state.Certificate = types.ListNull(types.StringType)
	} else {
		listVal, diags = types.ListValueFrom(ctx, types.StringType, apiResp.Certificate)
		resp.Diagnostics.Append(diags...)
		state.Certificate = listVal
	}

	state.Customer = types.StringPointerValue(apiResp.Customer)
	state.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *LbvServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan LbvServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state LbvServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := lbvServerAPIModel{
		Name: plan.Name.ValueString(),
	}

	var servicegroup []string
	resp.Diagnostics.Append(plan.Servicegroup.ElementsAs(ctx, &servicegroup, false)...)
	apiModel.Servicegroup = servicegroup

	if !plan.Ipaddress.IsNull() {
		v := plan.Ipaddress.ValueString()
		apiModel.Ipaddress = &v
	}
	if !plan.Port.IsNull() {
		v := plan.Port.ValueInt64()
		apiModel.Port = &v
	}
	if !plan.Type.IsNull() {
		v := plan.Type.ValueString()
		apiModel.Type = &v
	}
	if !plan.Certificate.IsNull() {
		var certificate []string
		resp.Diagnostics.Append(plan.Certificate.ElementsAs(ctx, &certificate, false)...)
		apiModel.Certificate = certificate
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp lbvServerAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/lbvserver/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating lbv_server", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	if apiResp.Ipaddress != nil {
		plan.Ipaddress = types.StringPointerValue(apiResp.Ipaddress)
	}
	plan.Port = types.Int64PointerValue(apiResp.Port)
	plan.Type = types.StringPointerValue(apiResp.Type)

	listVal, diags := types.ListValueFrom(ctx, types.StringType, apiResp.Servicegroup)
	resp.Diagnostics.Append(diags...)
	plan.Servicegroup = listVal

	if plan.Certificate.IsNull() && len(apiResp.Certificate) == 0 {
		plan.Certificate = types.ListNull(types.StringType)
	} else {
		listVal, diags = types.ListValueFrom(ctx, types.StringType, apiResp.Certificate)
		resp.Diagnostics.Append(diags...)
		plan.Certificate = listVal
	}

	plan.Customer = types.StringPointerValue(apiResp.Customer)
	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *LbvServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state LbvServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/lbvserver/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting lbv_server", err.Error())
	}
}
