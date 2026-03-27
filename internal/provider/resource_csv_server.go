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

var _ resource.Resource = &CsvServerResource{}

type CsvServerResource struct {
	client *apiclient.Client
}

type CsvServerResourceModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Ufname       types.String `tfsdk:"ufname"`
	Ipv46        types.String `tfsdk:"ipv46"`
	Port         types.Int64  `tfsdk:"port"`
	Type         types.String `tfsdk:"type"`
	Policies     types.List   `tfsdk:"policies"`
	Customer     types.String `tfsdk:"customer"`
	Certificate  types.List   `tfsdk:"certificate"`
	Loadbalancer types.String `tfsdk:"loadbalancer"`
}

type csvServerAPIModel struct {
	Id           string   `json:"id,omitempty"`
	Name         string   `json:"name"`
	Ufname       string   `json:"ufname"`
	Ipv46        *string  `json:"ipv46,omitempty"`
	Port         *int64   `json:"port,omitempty"`
	Type         string   `json:"type"`
	Policies     []string `json:"policies,omitempty"`
	Customer     *string  `json:"customer,omitempty"`
	Certificate  []string `json:"certificate,omitempty"`
	Loadbalancer *string  `json:"loadbalancer,omitempty"`
}

func NewCsvServerResource() resource.Resource {
	return &CsvServerResource{}
}

func (r *CsvServerResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_csv_server"
}

func (r *CsvServerResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"ufname": schema.StringAttribute{
				Required: true,
			},
			"ipv46": schema.StringAttribute{
				Optional: true,
			},
			"port": schema.Int64Attribute{
				Optional: true,
			},
			"type": schema.StringAttribute{
				Required: true,
			},
			"policies": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"customer": schema.StringAttribute{
				Optional: true,
			},
			"certificate": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"loadbalancer": schema.StringAttribute{
				Optional: true,
			},
		},
	}
}

func (r *CsvServerResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CsvServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CsvServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := csvServerAPIModel{
		Name:   plan.Name.ValueString(),
		Ufname: plan.Ufname.ValueString(),
		Type:   plan.Type.ValueString(),
	}
	if !plan.Ipv46.IsNull() {
		v := plan.Ipv46.ValueString()
		apiModel.Ipv46 = &v
	}
	if !plan.Port.IsNull() {
		v := plan.Port.ValueInt64()
		apiModel.Port = &v
	}
	if !plan.Policies.IsNull() {
		var policies []string
		resp.Diagnostics.Append(plan.Policies.ElementsAs(ctx, &policies, false)...)
		apiModel.Policies = policies
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Certificate.IsNull() {
		var certificate []string
		resp.Diagnostics.Append(plan.Certificate.ElementsAs(ctx, &certificate, false)...)
		apiModel.Certificate = certificate
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp csvServerAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/csvserver/", apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating csv_server", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Ufname = types.StringValue(apiResp.Ufname)
	plan.Ipv46 = types.StringPointerValue(apiResp.Ipv46)
	plan.Port = types.Int64PointerValue(apiResp.Port)
	plan.Type = types.StringValue(apiResp.Type)

	listVal, diags := types.ListValueFrom(ctx, types.StringType, apiResp.Policies)
	resp.Diagnostics.Append(diags...)
	plan.Policies = listVal

	plan.Customer = types.StringPointerValue(apiResp.Customer)

	listVal, diags = types.ListValueFrom(ctx, types.StringType, apiResp.Certificate)
	resp.Diagnostics.Append(diags...)
	plan.Certificate = listVal

	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CsvServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CsvServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp csvServerAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/csvserver/%s/", state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading csv_server", err.Error())
		return
	}

	state.Id = types.StringValue(apiResp.Id)
	state.Name = types.StringValue(apiResp.Name)
	state.Ufname = types.StringValue(apiResp.Ufname)
	state.Ipv46 = types.StringPointerValue(apiResp.Ipv46)
	state.Port = types.Int64PointerValue(apiResp.Port)
	state.Type = types.StringValue(apiResp.Type)

	listVal, diags := types.ListValueFrom(ctx, types.StringType, apiResp.Policies)
	resp.Diagnostics.Append(diags...)
	state.Policies = listVal

	state.Customer = types.StringPointerValue(apiResp.Customer)

	listVal, diags = types.ListValueFrom(ctx, types.StringType, apiResp.Certificate)
	resp.Diagnostics.Append(diags...)
	state.Certificate = listVal

	state.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *CsvServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CsvServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state CsvServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := csvServerAPIModel{
		Name:   plan.Name.ValueString(),
		Ufname: plan.Ufname.ValueString(),
		Type:   plan.Type.ValueString(),
	}
	if !plan.Ipv46.IsNull() {
		v := plan.Ipv46.ValueString()
		apiModel.Ipv46 = &v
	}
	if !plan.Port.IsNull() {
		v := plan.Port.ValueInt64()
		apiModel.Port = &v
	}
	if !plan.Policies.IsNull() {
		var policies []string
		resp.Diagnostics.Append(plan.Policies.ElementsAs(ctx, &policies, false)...)
		apiModel.Policies = policies
	}
	if !plan.Customer.IsNull() {
		v := plan.Customer.ValueString()
		apiModel.Customer = &v
	}
	if !plan.Certificate.IsNull() {
		var certificate []string
		resp.Diagnostics.Append(plan.Certificate.ElementsAs(ctx, &certificate, false)...)
		apiModel.Certificate = certificate
	}
	if !plan.Loadbalancer.IsNull() {
		v := plan.Loadbalancer.ValueString()
		apiModel.Loadbalancer = &v
	}

	var apiResp csvServerAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/csvserver/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating csv_server", err.Error())
		return
	}

	plan.Id = types.StringValue(apiResp.Id)
	plan.Name = types.StringValue(apiResp.Name)
	plan.Ufname = types.StringValue(apiResp.Ufname)
	plan.Ipv46 = types.StringPointerValue(apiResp.Ipv46)
	plan.Port = types.Int64PointerValue(apiResp.Port)
	plan.Type = types.StringValue(apiResp.Type)

	listVal, diags := types.ListValueFrom(ctx, types.StringType, apiResp.Policies)
	resp.Diagnostics.Append(diags...)
	plan.Policies = listVal

	plan.Customer = types.StringPointerValue(apiResp.Customer)

	listVal, diags = types.ListValueFrom(ctx, types.StringType, apiResp.Certificate)
	resp.Diagnostics.Append(diags...)
	plan.Certificate = listVal

	plan.Loadbalancer = types.StringPointerValue(apiResp.Loadbalancer)

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *CsvServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CsvServerResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("/api/loadbalancing/csvserver/%s/", state.Id.ValueString()))
	if err != nil {
		if apiclient.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error deleting csv_server", err.Error())
	}
}
