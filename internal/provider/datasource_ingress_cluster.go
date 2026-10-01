package provider

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &IngressClusterDataSource{}

type IngressClusterDataSource struct {
	client *apiclient.Client
}

type IngressClusterDataSourceModel struct {
	Id                 types.String          `tfsdk:"id"`
	Name               types.String          `tfsdk:"name"`
	Sla                types.String          `tfsdk:"sla"`
	Bandwidth          types.Int64           `tfsdk:"bandwidth"`
	Customer           types.String          `tfsdk:"customer"`
	Ipaddress          types.String          `tfsdk:"ipaddress"`
	Firewall           types.String          `tfsdk:"firewall"`
	Slug               types.String          `tfsdk:"slug"`
	State              types.String          `tfsdk:"state"`
	ReverseProxy       types.Int64           `tfsdk:"reverse_proxy"`
	ReverseProxyName   types.String          `tfsdk:"reverse_proxy_name"`
	IpaddressAddress   types.String          `tfsdk:"ipaddress_address"`
	IpaddressType      types.String          `tfsdk:"ipaddress_type"`
	CreatedAtTimestamp types.String          `tfsdk:"created_at_timestamp"`
	UpdatedAtTimestamp types.String          `tfsdk:"updated_at_timestamp"`
	CreatedByUser      types.Int64           `tfsdk:"created_by_user"`
	UpdatedByUser      types.Int64           `tfsdk:"updated_by_user"`
	IngressClusters    []IngressClusterModel `tfsdk:"ingress_clusters"`
}

// IngressClusterModel mirrors IngressClusterResourceModel field-for-field (allows a direct conversion).
type IngressClusterModel struct {
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

func NewIngressClusterDataSource() datasource.DataSource {
	return &IngressClusterDataSource{}
}

func (d *IngressClusterDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ingress_cluster"
}

func (d *IngressClusterDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	itemAttrs := map[string]schema.Attribute{
		"id":                   schema.StringAttribute{Computed: true},
		"name":                 schema.StringAttribute{Computed: true},
		"sla":                  schema.StringAttribute{Computed: true},
		"bandwidth":            schema.Int64Attribute{Computed: true},
		"customer":             schema.StringAttribute{Computed: true},
		"ipaddress":            schema.StringAttribute{Computed: true},
		"firewall":             schema.StringAttribute{Computed: true},
		"slug":                 schema.StringAttribute{Computed: true},
		"state":                schema.StringAttribute{Computed: true},
		"reverse_proxy":        schema.Int64Attribute{Computed: true},
		"reverse_proxy_name":   schema.StringAttribute{Computed: true},
		"ipaddress_address":    schema.StringAttribute{Computed: true},
		"ipaddress_type":       schema.StringAttribute{Computed: true},
		"created_at_timestamp": schema.StringAttribute{Computed: true},
		"updated_at_timestamp": schema.StringAttribute{Computed: true},
		"created_by_user":      schema.Int64Attribute{Computed: true},
		"updated_by_user":      schema.Int64Attribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up secure ingress clusters. Set `name` or `id` to fetch a single cluster; otherwise all clusters matching the optional `ipaddress`, `sla` and `bandwidth` filters are listed.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific ingress cluster to look up.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Exact ingress cluster name to look up.",
			},
			"sla": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Filter by service level (bronze, silver, gold, platinum); the cluster's SLA for a single lookup.",
				Validators:  []validator.String{stringvalidator.OneOf(ingressClusterSLAs...)},
			},
			"bandwidth": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Filter by bandwidth in Mbps (50, 100, 500, 1000, 5000); the cluster's bandwidth for a single lookup.",
				Validators:  []validator.Int64{int64validator.OneOf(ingressClusterBandwidths...)},
			},
			"ipaddress": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Filter by public IP address UUID; the cluster's public IP address UUID for a single lookup.",
			},
			"customer":             schema.StringAttribute{Computed: true, Description: "Customer identifier."},
			"firewall":             schema.StringAttribute{Computed: true, Description: "UUID of the secure ingress firewall."},
			"slug":                 schema.StringAttribute{Computed: true, Description: "URL-safe identifier derived by the API."},
			"state":                schema.StringAttribute{Computed: true, Description: "Synchronisation state: synced, unsynced, error or deleted."},
			"reverse_proxy":        schema.Int64Attribute{Computed: true, Description: "ID of the reverse proxy integration."},
			"reverse_proxy_name":   schema.StringAttribute{Computed: true, Description: "Name of the reverse proxy integration."},
			"ipaddress_address":    schema.StringAttribute{Computed: true, Description: "The public IP address the cluster listens on."},
			"ipaddress_type":       schema.StringAttribute{Computed: true, Description: "Type of the public IP address."},
			"created_at_timestamp": schema.StringAttribute{Computed: true, Description: "Time when the cluster was created."},
			"updated_at_timestamp": schema.StringAttribute{Computed: true, Description: "Time when the cluster was last updated."},
			"created_by_user":      schema.Int64Attribute{Computed: true, Description: "ID of the user who created the cluster."},
			"updated_by_user":      schema.Int64Attribute{Computed: true, Description: "ID of the user who last updated the cluster."},
			"ingress_clusters": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All matching ingress clusters (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: itemAttrs},
			},
		},
	}
}

func (d *IngressClusterDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected provider data type",
			fmt.Sprintf("Expected *apiclient.Client, got %T", req.ProviderData))
		return
	}
	d.client = client
}

func (d *IngressClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config IngressClusterDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item ingressClusterAPIModel
		if err := d.client.Get(ctx, fmt.Sprintf("%s%s/", ingressClusterBasePath, config.Id.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Error reading ingress cluster", err.Error())
			return
		}
		state := ingressClusterSingleState(&item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	query := url.Values{}
	name := config.Name.ValueString()
	if !config.Name.IsNull() && name != "" {
		query.Set("name__icontains", name)
	}
	if !config.Ipaddress.IsNull() && config.Ipaddress.ValueString() != "" {
		query.Set("ipaddress", config.Ipaddress.ValueString())
	}
	if !config.Sla.IsNull() && config.Sla.ValueString() != "" {
		query.Set("sla", config.Sla.ValueString())
	}
	if !config.Bandwidth.IsNull() {
		query.Set("bandwidth", strconv.FormatInt(config.Bandwidth.ValueInt64(), 10))
	}
	listPath := ingressClusterBasePath
	if len(query) > 0 {
		listPath += "?" + query.Encode()
	}

	items, err := listAll[ingressClusterAPIModel](ctx, d.client, listPath)
	if err != nil {
		resp.Diagnostics.AddError("Error listing ingress clusters", err.Error())
		return
	}

	if !config.Name.IsNull() && name != "" {
		for i := range items {
			if items[i].Name == name {
				state := ingressClusterSingleState(&items[i])
				resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
				return
			}
		}
		resp.Diagnostics.AddError("Ingress cluster not found",
			fmt.Sprintf("No ingress cluster with exact name %q was found.", name))
		return
	}

	state := IngressClusterDataSourceModel{
		Id:                 types.StringNull(),
		Name:               types.StringNull(),
		Sla:                config.Sla,
		Bandwidth:          config.Bandwidth,
		Customer:           types.StringNull(),
		Ipaddress:          config.Ipaddress,
		Firewall:           types.StringNull(),
		Slug:               types.StringNull(),
		State:              types.StringNull(),
		ReverseProxy:       types.Int64Null(),
		ReverseProxyName:   types.StringNull(),
		IpaddressAddress:   types.StringNull(),
		IpaddressType:      types.StringNull(),
		CreatedAtTimestamp: types.StringNull(),
		UpdatedAtTimestamp: types.StringNull(),
		CreatedByUser:      types.Int64Null(),
		UpdatedByUser:      types.Int64Null(),
		IngressClusters:    make([]IngressClusterModel, 0, len(items)),
	}
	for i := range items {
		state.IngressClusters = append(state.IngressClusters, ingressClusterItemToModel(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func ingressClusterItemToModel(a *ingressClusterAPIModel) IngressClusterModel {
	var m IngressClusterResourceModel
	mapIngressClusterToState(&m, a)
	return IngressClusterModel(m)
}

func ingressClusterSingleState(a *ingressClusterAPIModel) IngressClusterDataSourceModel {
	m := ingressClusterItemToModel(a)
	return IngressClusterDataSourceModel{
		Id:                 m.Id,
		Name:               m.Name,
		Sla:                m.Sla,
		Bandwidth:          m.Bandwidth,
		Customer:           m.Customer,
		Ipaddress:          m.Ipaddress,
		Firewall:           m.Firewall,
		Slug:               m.Slug,
		State:              m.State,
		ReverseProxy:       m.ReverseProxy,
		ReverseProxyName:   m.ReverseProxyName,
		IpaddressAddress:   m.IpaddressAddress,
		IpaddressType:      m.IpaddressType,
		CreatedAtTimestamp: m.CreatedAtTimestamp,
		UpdatedAtTimestamp: m.UpdatedAtTimestamp,
		CreatedByUser:      m.CreatedByUser,
		UpdatedByUser:      m.UpdatedByUser,
		IngressClusters:    []IngressClusterModel{},
	}
}
