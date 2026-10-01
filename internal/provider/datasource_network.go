package provider

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NetworkDataSource{}

type NetworkDataSource struct {
	client *apiclient.Client
}

type NetworkDataSourceModel struct {
	Name         types.String   `tfsdk:"name"`
	Id           types.String   `tfsdk:"id"`
	Domain       types.Int64    `tfsdk:"domain"`
	DomainName   types.String   `tfsdk:"domain_name"`
	DomainVdomId types.String   `tfsdk:"domain_vdom_id"`
	Description  types.String   `tfsdk:"description"`
	Ipv4Prefix   types.String   `tfsdk:"ipv4_prefix"`
	Ipv4Address  types.String   `tfsdk:"ipv4_address"`
	VlanId       types.Int64    `tfsdk:"vlan_id"`
	Networks     []NetworkModel `tfsdk:"networks"`
}

type NetworkModel struct {
	Id           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Domain       types.Int64  `tfsdk:"domain"`
	DomainName   types.String `tfsdk:"domain_name"`
	DomainVdomId types.String `tfsdk:"domain_vdom_id"`
	Description  types.String `tfsdk:"description"`
	Ipv4Prefix   types.String `tfsdk:"ipv4_prefix"`
	Ipv4Address  types.String `tfsdk:"ipv4_address"`
	VlanId       types.Int64  `tfsdk:"vlan_id"`
}

type networkAPIModel struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Domain       *int64 `json:"domain"`
	DomainDetail *struct {
		Id     int64  `json:"id"`
		VdomId string `json:"vdom_id"`
		Name   string `json:"name"`
	} `json:"domain_detail"`
	Description string `json:"description"`
	Ipv4Prefix  string `json:"ipv4_prefix"`
	Ipv4Address string `json:"ipv4_address"`
	VlanId      *int64 `json:"vlanid"`
}

func (n *networkAPIModel) toModel() NetworkModel {
	m := NetworkModel{
		Id:           types.StringValue(n.Id),
		Name:         types.StringValue(n.Name),
		Domain:       types.Int64PointerValue(n.Domain),
		DomainName:   types.StringNull(),
		DomainVdomId: types.StringNull(),
		Description:  types.StringValue(n.Description),
		Ipv4Prefix:   types.StringValue(n.Ipv4Prefix),
		Ipv4Address:  types.StringValue(n.Ipv4Address),
		VlanId:       types.Int64PointerValue(n.VlanId),
	}
	if n.DomainDetail != nil {
		m.DomainName = types.StringValue(n.DomainDetail.Name)
		m.DomainVdomId = types.StringValue(n.DomainDetail.VdomId)
	}
	return m
}

func NewNetworkDataSource() datasource.DataSource {
	return &NetworkDataSource{}
}

func (d *NetworkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (d *NetworkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	networkAttrs := map[string]schema.Attribute{
		"id":             schema.StringAttribute{Computed: true},
		"name":           schema.StringAttribute{Computed: true},
		"domain":         schema.Int64Attribute{Computed: true},
		"domain_name":    schema.StringAttribute{Computed: true},
		"domain_vdom_id": schema.StringAttribute{Computed: true},
		"description":    schema.StringAttribute{Computed: true},
		"ipv4_prefix":    schema.StringAttribute{Computed: true},
		"ipv4_address":   schema.StringAttribute{Computed: true},
		"vlan_id":        schema.Int64Attribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS networks. Set `name` to fetch a single network by exact name, or omit it to list all networks (optionally filtered by `domain`).",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact network name to look up. When set, the data source returns a single network and the `networks` list is empty.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the matched network (only set when `name` is provided).",
			},
			"domain": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Filter by domain ID; set to the domain of the matched network.",
			},
			"domain_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the matched network's domain.",
			},
			"domain_vdom_id": schema.StringAttribute{
				Computed:    true,
				Description: "VDOM ID of the matched network's domain.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the matched network.",
			},
			"ipv4_prefix": schema.StringAttribute{
				Computed:    true,
				Description: "IPv4 prefix of the matched network (only set when `name` is provided).",
			},
			"ipv4_address": schema.StringAttribute{
				Computed:    true,
				Description: "IPv4 address of the matched network.",
			},
			"vlan_id": schema.Int64Attribute{
				Computed:    true,
				Description: "VLAN ID of the matched network (only set when `name` is provided).",
			},
			"networks": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All networks (populated when `name` is not set).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: networkAttrs,
				},
			},
		},
	}
}

func (d *NetworkDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetworkDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config NetworkDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	q := url.Values{}
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		q.Set("name__icontains", config.Name.ValueString())
	}
	if !config.Domain.IsNull() && !config.Domain.IsUnknown() {
		q.Set("domain", strconv.FormatInt(config.Domain.ValueInt64(), 10))
	}
	path := "/api/networking/networks/"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	items, err := listAll[networkAPIModel](ctx, d.client, path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading networks", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *networkAPIModel
		for i := range items {
			if items[i].Name == config.Name.ValueString() {
				match = &items[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Network not found",
				fmt.Sprintf("No network with exact name %q was found.", config.Name.ValueString()))
			return
		}
		m := match.toModel()
		config.Id = m.Id
		config.Domain = m.Domain
		config.DomainName = m.DomainName
		config.DomainVdomId = m.DomainVdomId
		config.Description = m.Description
		config.Ipv4Prefix = m.Ipv4Prefix
		config.Ipv4Address = m.Ipv4Address
		config.VlanId = m.VlanId
		config.Networks = []NetworkModel{}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := NetworkDataSourceModel{
		Name:         types.StringNull(),
		Id:           types.StringNull(),
		Domain:       config.Domain,
		DomainName:   types.StringNull(),
		DomainVdomId: types.StringNull(),
		Description:  types.StringNull(),
		Ipv4Prefix:   types.StringNull(),
		Ipv4Address:  types.StringNull(),
		VlanId:       types.Int64Null(),
		Networks:     make([]NetworkModel, 0, len(items)),
	}
	if state.Domain.IsUnknown() {
		state.Domain = types.Int64Null()
	}
	for i := range items {
		state.Networks = append(state.Networks, items[i].toModel())
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
