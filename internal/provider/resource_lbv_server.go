package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &LbvServerResource{}
	_ resource.ResourceWithImportState = &LbvServerResource{}
)

type LbvServerResource struct {
	client *apiclient.Client
}

type LbvServerResourceModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Ipaddress     types.String `tfsdk:"ipaddress"`
	Port          types.Int64  `tfsdk:"port"`
	Type          types.String `tfsdk:"type"`
	Servicegroup  types.List   `tfsdk:"servicegroup"`
	Certificate   types.List   `tfsdk:"certificate"`
	CaCertificate types.List   `tfsdk:"ca_certificate"`
	Customer      types.String `tfsdk:"customer"`
	Loadbalancer  types.String `tfsdk:"loadbalancer"`
}

type lbvServerAPIModel struct {
	Id            string    `json:"id,omitempty"`
	Name          string    `json:"name"`
	Ipaddress     *string   `json:"ipaddress"`
	Port          *int64    `json:"port,omitempty"`
	Type          *string   `json:"type,omitempty"`
	Servicegroup  []string  `json:"servicegroup"`
	Certificate   *[]string `json:"certificate,omitempty"`
	CaCertificate *[]string `json:"ca_certificate,omitempty"`
	Customer      *string   `json:"customer,omitempty"`
	Loadbalancer  *string   `json:"loadbalancer,omitempty"`
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
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Port number. Leave at the default (0) if used as non routed loadbalancer.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ssl"),
				Description: "One of `http`, `ssl`, `ssl_bridge`, `tcp`, `udp`. Defaults to `ssl`.",
				Validators:  []validator.String{stringvalidator.OneOf("http", "ssl", "ssl_bridge", "tcp", "udp")},
			},
			"servicegroup": schema.ListAttribute{
				ElementType: types.StringType,
				Required:    true,
			},
			"certificate": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
			},
			"ca_certificate": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "UUIDs of CA certificates.",
			},
			"customer": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"loadbalancer": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
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

func lbvServerToAPI(ctx context.Context, plan, state *LbvServerResourceModel, diags *diag.Diagnostics) lbvServerAPIModel {
	m := lbvServerAPIModel{
		Name:         plan.Name.ValueString(),
		Ipaddress:    stringPtr(plan.Ipaddress),
		Port:         int64Ptr(plan.Port),
		Type:         stringPtr(plan.Type),
		Servicegroup: []string{},
		Customer:     stringPtr(plan.Customer),
		Loadbalancer: stringPtr(plan.Loadbalancer),
	}
	diags.Append(plan.Servicegroup.ElementsAs(ctx, &m.Servicegroup, false)...)
	if state == nil {
		m.Certificate = listElems[string](ctx, plan.Certificate, diags)
		m.CaCertificate = listElems[string](ctx, plan.CaCertificate, diags)
	} else {
		m.Certificate = listElemsForUpdate[string](ctx, plan.Certificate, state.Certificate, diags)
		m.CaCertificate = listElemsForUpdate[string](ctx, plan.CaCertificate, state.CaCertificate, diags)
	}
	return m
}

func lbvServerFromAPI(ctx context.Context, m *LbvServerResourceModel, api *lbvServerAPIModel, diags *diag.Diagnostics) {
	m.Id = types.StringValue(api.Id)
	m.Name = types.StringValue(api.Name)
	m.Ipaddress = types.StringPointerValue(api.Ipaddress)
	// port is nullable in the API; null is equivalent to the default 0 (non routed loadbalancer).
	if api.Port != nil {
		m.Port = types.Int64Value(*api.Port)
	} else {
		m.Port = types.Int64Value(0)
	}
	if api.Type != nil {
		m.Type = types.StringValue(*api.Type)
	} else if m.Type.IsUnknown() {
		m.Type = types.StringNull()
	}
	m.Servicegroup = computedListValue(ctx, types.StringType, api.Servicegroup, diags)
	m.Certificate = listValue(ctx, types.StringType, m.Certificate, lbStringSlice(api.Certificate), diags)
	m.CaCertificate = listValue(ctx, types.StringType, m.CaCertificate, lbStringSlice(api.CaCertificate), diags)
	m.Customer = types.StringPointerValue(api.Customer)
	m.Loadbalancer = types.StringPointerValue(api.Loadbalancer)
}

func (r *LbvServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan LbvServerResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	apiModel := lbvServerToAPI(ctx, &plan, nil, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbvServerAPIModel
	err := r.client.Post(ctx, "/api/loadbalancing/lbvserver/", apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error creating lbv_server", err.Error())
		return
	}

	lbvServerFromAPI(ctx, &plan, &apiResp, &resp.Diagnostics)
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

	lbvServerFromAPI(ctx, &state, &apiResp, &resp.Diagnostics)
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

	apiModel := lbvServerToAPI(ctx, &plan, &state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp lbvServerAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("/api/loadbalancing/lbvserver/%s/", state.Id.ValueString()), apiModel, &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating lbv_server", err.Error())
		return
	}

	lbvServerFromAPI(ctx, &plan, &apiResp, &resp.Diagnostics)
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

func (r *LbvServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
