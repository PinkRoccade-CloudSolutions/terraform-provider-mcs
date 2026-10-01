package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const ingressClusterBasePath = "/api/secureingress/ingresscluster/"

var (
	_ resource.Resource                = &IngressClusterResource{}
	_ resource.ResourceWithImportState = &IngressClusterResource{}
)

// ingressClusterSLAs and ingressClusterBandwidths are the spec's SlaEnum and BandwidthEnum (Mbps).
var (
	ingressClusterSLAs       = []string{"bronze", "silver", "gold", "platinum"}
	ingressClusterBandwidths = []int64{50, 100, 500, 1000, 5000}
)

type IngressClusterResource struct {
	client *apiclient.Client
}

type IngressClusterResourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Sla                types.String `tfsdk:"sla"`
	Bandwidth          types.Int64  `tfsdk:"bandwidth"`
	Customer           types.String `tfsdk:"customer"`
	Ipaddress          types.String `tfsdk:"ipaddress"`
	Firewall           types.String `tfsdk:"firewall"`
	Slug               types.String `tfsdk:"slug"`
	State              types.String `tfsdk:"state"`
	ReverseProxy       types.Int64  `tfsdk:"reverse_proxy"`
	ReverseProxyName   types.String `tfsdk:"reverse_proxy_name"`
	IpaddressAddress   types.String `tfsdk:"ipaddress_address"`
	IpaddressType      types.String `tfsdk:"ipaddress_type"`
	CreatedAtTimestamp types.String `tfsdk:"created_at_timestamp"`
	UpdatedAtTimestamp types.String `tfsdk:"updated_at_timestamp"`
	CreatedByUser      types.Int64  `tfsdk:"created_by_user"`
	UpdatedByUser      types.Int64  `tfsdk:"updated_by_user"`
}

// ingressClusterRequest is the writable subset of the IngressCluster schema.
type ingressClusterRequest struct {
	Name      string  `json:"name"`
	Sla       *string `json:"sla,omitempty"`
	Bandwidth *int64  `json:"bandwidth,omitempty"`
	Customer  string  `json:"customer"`
	Ipaddress string  `json:"ipaddress"`
	Firewall  string  `json:"firewall"`
}

type ingressClusterAPIModel struct {
	Id                 string  `json:"id"`
	Name               string  `json:"name"`
	Slug               string  `json:"slug"`
	Sla                *string `json:"sla"`
	Bandwidth          *int64  `json:"bandwidth"`
	State              string  `json:"state"`
	Customer           string  `json:"customer"`
	Ipaddress          string  `json:"ipaddress"`
	Firewall           string  `json:"firewall"`
	ReverseProxy       *int64  `json:"reverse_proxy"`
	CreatedAtTimestamp string  `json:"created_at_timestamp"`
	UpdatedAtTimestamp string  `json:"updated_at_timestamp"`
	CreatedByUser      *int64  `json:"created_by_user"`
	UpdatedByUser      *int64  `json:"updated_by_user"`
	ReverseProxyDetail *struct {
		Id   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"reverse_proxy_detail"`
	IpaddressDetail *struct {
		Id      string `json:"id"`
		Address string `json:"address"`
		Type    string `json:"type"`
	} `json:"ipaddress_detail"`
}

func NewIngressClusterResource() resource.Resource {
	return &IngressClusterResource{}
}

func (r *IngressClusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ingress_cluster"
}

func (r *IngressClusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a secure ingress cluster: a public IP address protected by a secure ingress (XDP) firewall and fronted by a reverse proxy.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:      true,
				Description:   "UUID of the ingress cluster.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Name of the ingress cluster (max 200 characters).",
				Validators:  []validator.String{stringvalidator.LengthBetween(1, 200)},
			},
			"sla": schema.StringAttribute{
				Optional:      true,
				Computed:      true,
				Description:   "Service level: `bronze`, `silver`, `gold` or `platinum`. Defaults to the API's default when unset.",
				Validators:    []validator.String{stringvalidator.OneOf(ingressClusterSLAs...)},
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"bandwidth": schema.Int64Attribute{
				Optional:      true,
				Computed:      true,
				Description:   "Bandwidth in Mbps: `50`, `100`, `500`, `1000` or `5000`. Defaults to the API's default when unset.",
				Validators:    []validator.Int64{int64validator.OneOf(ingressClusterBandwidths...)},
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"customer": schema.StringAttribute{
				Required:      true,
				Description:   "Customer identifier. Changing this forces a new resource.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"ipaddress": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the public IP address (`mcs_public_ip_address`) the cluster listens on.",
			},
			"firewall": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the secure ingress firewall (`mcs_secureingress_firewall`) applied to the cluster.",
			},
			"slug": schema.StringAttribute{
				Computed:    true,
				Description: "URL-safe identifier derived by the API.",
			},
			"state": schema.StringAttribute{
				Computed:    true,
				Description: "Synchronisation state: synced, unsynced, error or deleted.",
			},
			"reverse_proxy": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the reverse proxy integration serving the cluster.",
			},
			"reverse_proxy_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the reverse proxy integration serving the cluster.",
			},
			"ipaddress_address": schema.StringAttribute{
				Computed:    true,
				Description: "The public IP address the cluster listens on.",
			},
			"ipaddress_type": schema.StringAttribute{
				Computed:    true,
				Description: "Type of the public IP address.",
			},
			"created_at_timestamp": schema.StringAttribute{
				Computed:      true,
				Description:   "Time when the cluster was created.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"updated_at_timestamp": schema.StringAttribute{
				Computed:    true,
				Description: "Time when the cluster was last updated.",
			},
			"created_by_user": schema.Int64Attribute{
				Computed:      true,
				Description:   "ID of the user who created the cluster.",
				PlanModifiers: []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"updated_by_user": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the user who last updated the cluster.",
			},
		},
	}
}

func (r *IngressClusterResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *apiclient.Client, got %T", req.ProviderData))
		return
	}
	r.client = client
}

func buildIngressClusterRequest(plan *IngressClusterResourceModel) ingressClusterRequest {
	return ingressClusterRequest{
		Name:      plan.Name.ValueString(),
		Sla:       stringPtr(plan.Sla),
		Bandwidth: int64Ptr(plan.Bandwidth),
		Customer:  plan.Customer.ValueString(),
		Ipaddress: plan.Ipaddress.ValueString(),
		Firewall:  plan.Firewall.ValueString(),
	}
}

func mapIngressClusterToState(m *IngressClusterResourceModel, a *ingressClusterAPIModel) {
	m.Id = types.StringValue(a.Id)
	m.Name = types.StringValue(a.Name)
	m.Sla = types.StringPointerValue(a.Sla)
	m.Bandwidth = types.Int64PointerValue(a.Bandwidth)
	m.Customer = types.StringValue(a.Customer)
	m.Ipaddress = types.StringValue(a.Ipaddress)
	m.Firewall = types.StringValue(a.Firewall)
	m.Slug = types.StringValue(a.Slug)
	m.State = types.StringValue(a.State)
	m.ReverseProxy = types.Int64PointerValue(a.ReverseProxy)
	m.ReverseProxyName = types.StringNull()
	if a.ReverseProxyDetail != nil {
		m.ReverseProxyName = types.StringValue(a.ReverseProxyDetail.Name)
	}
	m.IpaddressAddress = types.StringNull()
	m.IpaddressType = types.StringNull()
	if a.IpaddressDetail != nil {
		m.IpaddressAddress = types.StringValue(a.IpaddressDetail.Address)
		m.IpaddressType = types.StringValue(a.IpaddressDetail.Type)
	}
	m.CreatedAtTimestamp = types.StringValue(a.CreatedAtTimestamp)
	m.UpdatedAtTimestamp = types.StringValue(a.UpdatedAtTimestamp)
	m.CreatedByUser = types.Int64PointerValue(a.CreatedByUser)
	m.UpdatedByUser = types.Int64PointerValue(a.UpdatedByUser)
}

func (r *IngressClusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan IngressClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp ingressClusterAPIModel
	if err := r.client.Post(ctx, ingressClusterBasePath, buildIngressClusterRequest(&plan), &apiResp); err != nil {
		resp.Diagnostics.AddError("Error creating ingress cluster", err.Error())
		return
	}

	mapIngressClusterToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IngressClusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state IngressClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp ingressClusterAPIModel
	err := r.client.Get(ctx, fmt.Sprintf("%s%s/", ingressClusterBasePath, state.Id.ValueString()), &apiResp)
	if err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading ingress cluster", err.Error())
		return
	}

	mapIngressClusterToState(&state, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *IngressClusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state IngressClusterResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var apiResp ingressClusterAPIModel
	err := r.client.Put(ctx, fmt.Sprintf("%s%s/", ingressClusterBasePath, state.Id.ValueString()), buildIngressClusterRequest(&plan), &apiResp)
	if err != nil {
		resp.Diagnostics.AddError("Error updating ingress cluster", err.Error())
		return
	}

	mapIngressClusterToState(&plan, &apiResp)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *IngressClusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state IngressClusterResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, fmt.Sprintf("%s%s/", ingressClusterBasePath, state.Id.ValueString()))
	if err != nil && !apiclient.IsNotFound(err) {
		resp.Diagnostics.AddError("Error deleting ingress cluster", err.Error())
	}
}

func (r *IngressClusterResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
