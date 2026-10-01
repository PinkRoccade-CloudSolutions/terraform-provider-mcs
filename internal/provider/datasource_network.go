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
	Name                  types.String   `tfsdk:"name"`
	Id                    types.String   `tfsdk:"id"`
	Domain                types.Int64    `tfsdk:"domain"`
	DomainName            types.String   `tfsdk:"domain_name"`
	DomainVdomId          types.String   `tfsdk:"domain_vdom_id"`
	Description           types.String   `tfsdk:"description"`
	Ipv4Prefix            types.String   `tfsdk:"ipv4_prefix"`
	Ipv4Address           types.String   `tfsdk:"ipv4_address"`
	VlanId                types.Int64    `tfsdk:"vlan_id"`
	Customer              types.String   `tfsdk:"customer"`
	CustomerName          types.String   `tfsdk:"customer_name"`
	Type                  types.String   `tfsdk:"type"`
	Parent                types.String   `tfsdk:"parent"`
	Ipv6Prefix            types.String   `tfsdk:"ipv6_prefix"`
	Ipv6Address           types.String   `tfsdk:"ipv6_address"`
	DhcpServer            types.Bool     `tfsdk:"dhcp_server"`
	DhcpServerAddressPool types.String   `tfsdk:"dhcp_server_address_pool"`
	DhcpServerNetmask     types.String   `tfsdk:"dhcp_server_netmask"`
	DhcpServerDnsServers  types.String   `tfsdk:"dhcp_server_dns_servers"`
	DhcpServerLeaseTime   types.Int64    `tfsdk:"dhcp_server_lease_time"`
	Visible               types.Bool     `tfsdk:"visible"`
	CreatedAt             types.String   `tfsdk:"created_at"`
	UpdatedAt             types.String   `tfsdk:"updated_at"`
	Networks              []NetworkModel `tfsdk:"networks"`
}

type NetworkModel struct {
	Id                    types.String `tfsdk:"id"`
	Name                  types.String `tfsdk:"name"`
	Domain                types.Int64  `tfsdk:"domain"`
	DomainName            types.String `tfsdk:"domain_name"`
	DomainVdomId          types.String `tfsdk:"domain_vdom_id"`
	Description           types.String `tfsdk:"description"`
	Ipv4Prefix            types.String `tfsdk:"ipv4_prefix"`
	Ipv4Address           types.String `tfsdk:"ipv4_address"`
	VlanId                types.Int64  `tfsdk:"vlan_id"`
	Customer              types.String `tfsdk:"customer"`
	CustomerName          types.String `tfsdk:"customer_name"`
	Type                  types.String `tfsdk:"type"`
	Parent                types.String `tfsdk:"parent"`
	Ipv6Prefix            types.String `tfsdk:"ipv6_prefix"`
	Ipv6Address           types.String `tfsdk:"ipv6_address"`
	DhcpServer            types.Bool   `tfsdk:"dhcp_server"`
	DhcpServerAddressPool types.String `tfsdk:"dhcp_server_address_pool"`
	DhcpServerNetmask     types.String `tfsdk:"dhcp_server_netmask"`
	DhcpServerDnsServers  types.String `tfsdk:"dhcp_server_dns_servers"`
	DhcpServerLeaseTime   types.Int64  `tfsdk:"dhcp_server_lease_time"`
	Visible               types.Bool   `tfsdk:"visible"`
	CreatedAt             types.String `tfsdk:"created_at"`
	UpdatedAt             types.String `tfsdk:"updated_at"`
}

// networkV3APIModel is the full Network record returned by /api/v3/networking/networks/ and
// embedded in networking operations. It is shared by the mcs_network data source and resource.
type networkV3APIModel struct {
	Id           string `json:"id"`
	Name         string `json:"name"`
	Domain       *int64 `json:"domain"`
	DomainDetail *struct {
		Id     int64  `json:"id"`
		VdomId string `json:"vdom_id"`
		Name   string `json:"name"`
	} `json:"domain_detail"`
	Customer              *string `json:"customer"`
	CustomerName          string  `json:"customer_name"`
	Description           string  `json:"description"`
	Type                  string  `json:"type"`
	Parent                string  `json:"parent"`
	VlanId                *int64  `json:"vlanid"`
	Ipv4Prefix            string  `json:"ipv4_prefix"`
	Ipv4Address           string  `json:"ipv4_address"`
	Ipv6Prefix            string  `json:"ipv6_prefix"`
	Ipv6Address           string  `json:"ipv6_address"`
	DhcpServer            *bool   `json:"dhcp_server"`
	DhcpServerAddressPool string  `json:"dhcp_server_address_pool"`
	DhcpServerNetmask     string  `json:"dhcp_server_netmask"`
	DhcpServerDnsServers  string  `json:"dhcp_server_dns_servers"`
	DhcpServerLeaseTime   *int64  `json:"dhcp_server_lease_time"`
	Visible               *bool   `json:"visible"`
	CreatedAt             string  `json:"created_at_timestamp"`
	UpdatedAt             string  `json:"updated_at_timestamp"`
}

func (n *networkV3APIModel) toModel() NetworkModel {
	m := NetworkModel{
		Id:                    types.StringValue(n.Id),
		Name:                  types.StringValue(n.Name),
		Domain:                types.Int64PointerValue(n.Domain),
		DomainName:            types.StringNull(),
		DomainVdomId:          types.StringNull(),
		Description:           types.StringValue(n.Description),
		Ipv4Prefix:            types.StringValue(n.Ipv4Prefix),
		Ipv4Address:           types.StringValue(n.Ipv4Address),
		VlanId:                types.Int64PointerValue(n.VlanId),
		Customer:              types.StringPointerValue(n.Customer),
		CustomerName:          types.StringValue(n.CustomerName),
		Type:                  types.StringValue(n.Type),
		Parent:                types.StringValue(n.Parent),
		Ipv6Prefix:            types.StringValue(n.Ipv6Prefix),
		Ipv6Address:           types.StringValue(n.Ipv6Address),
		DhcpServer:            types.BoolPointerValue(n.DhcpServer),
		DhcpServerAddressPool: types.StringValue(n.DhcpServerAddressPool),
		DhcpServerNetmask:     types.StringValue(n.DhcpServerNetmask),
		DhcpServerDnsServers:  types.StringValue(n.DhcpServerDnsServers),
		DhcpServerLeaseTime:   types.Int64PointerValue(n.DhcpServerLeaseTime),
		Visible:               types.BoolPointerValue(n.Visible),
		CreatedAt:             types.StringValue(n.CreatedAt),
		UpdatedAt:             types.StringValue(n.UpdatedAt),
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

// networkDetailAttrDescriptions describes the v3 fields shared by the top-level (single network)
// attributes and the entries of the `networks` list.
var networkDetailAttrDescriptions = map[string]string{
	"customer":                 "Customer that owns the network.",
	"customer_name":            "Name of the customer that owns the network.",
	"type":                     "Interface type: `vlan`, `tunnel`, `lacp` or `physical`.",
	"parent":                   "Parent LACP interface.",
	"ipv6_prefix":              "IPv6 prefix of the network.",
	"ipv6_address":             "IPv6 address of the network.",
	"dhcp_server":              "Whether a DHCP server is enabled on the interface.",
	"dhcp_server_address_pool": "DHCP server address pool.",
	"dhcp_server_netmask":      "DHCP server netmask.",
	"dhcp_server_dns_servers":  "Comma-separated DNS servers handed out by the DHCP server.",
	"dhcp_server_lease_time":   "DHCP lease time.",
	"visible":                  "Whether the interface is visible to end users.",
	"created_at":               "Time the network was created.",
	"updated_at":               "Time the network was last updated.",
}

func networkDetailAttrs(withDescriptions bool) map[string]schema.Attribute {
	desc := func(k string) string {
		if withDescriptions {
			return networkDetailAttrDescriptions[k]
		}
		return ""
	}
	return map[string]schema.Attribute{
		"customer":                 schema.StringAttribute{Computed: true, Description: desc("customer")},
		"customer_name":            schema.StringAttribute{Computed: true, Description: desc("customer_name")},
		"type":                     schema.StringAttribute{Computed: true, Description: desc("type")},
		"parent":                   schema.StringAttribute{Computed: true, Description: desc("parent")},
		"ipv6_prefix":              schema.StringAttribute{Computed: true, Description: desc("ipv6_prefix")},
		"ipv6_address":             schema.StringAttribute{Computed: true, Description: desc("ipv6_address")},
		"dhcp_server":              schema.BoolAttribute{Computed: true, Description: desc("dhcp_server")},
		"dhcp_server_address_pool": schema.StringAttribute{Computed: true, Description: desc("dhcp_server_address_pool")},
		"dhcp_server_netmask":      schema.StringAttribute{Computed: true, Description: desc("dhcp_server_netmask")},
		"dhcp_server_dns_servers":  schema.StringAttribute{Computed: true, Description: desc("dhcp_server_dns_servers")},
		"dhcp_server_lease_time":   schema.Int64Attribute{Computed: true, Description: desc("dhcp_server_lease_time")},
		"visible":                  schema.BoolAttribute{Computed: true, Description: desc("visible")},
		"created_at":               schema.StringAttribute{Computed: true, Description: desc("created_at")},
		"updated_at":               schema.StringAttribute{Computed: true, Description: desc("updated_at")},
	}
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
	for k, v := range networkDetailAttrs(false) {
		networkAttrs[k] = v
	}

	attrs := map[string]schema.Attribute{
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
	}
	for k, v := range networkDetailAttrs(true) {
		attrs[k] = v
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS networks. Set `name` to fetch a single network by exact name, or omit it to list all networks (optionally filtered by `domain`).",
		Attributes:  attrs,
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
	path := "/api/v3/networking/networks/"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	items, err := listAll[networkV3APIModel](ctx, d.client, path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading networks", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *networkV3APIModel
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
		state := networkDataSourceFromModel(m)
		state.Name = config.Name
		state.Networks = []NetworkModel{}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	state := networkDataSourceFromModel(NetworkModel{})
	state.Name = types.StringNull()
	state.Domain = config.Domain
	if state.Domain.IsUnknown() {
		state.Domain = types.Int64Null()
	}
	state.Networks = make([]NetworkModel, 0, len(items))
	for i := range items {
		state.Networks = append(state.Networks, items[i].toModel())
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// networkDataSourceFromModel copies a single network into the top-level attributes. A zero
// NetworkModel (all attributes null) leaves them null, which is what list mode wants.
func networkDataSourceFromModel(m NetworkModel) NetworkDataSourceModel {
	return NetworkDataSourceModel{
		Id:                    m.Id,
		Domain:                m.Domain,
		DomainName:            m.DomainName,
		DomainVdomId:          m.DomainVdomId,
		Description:           m.Description,
		Ipv4Prefix:            m.Ipv4Prefix,
		Ipv4Address:           m.Ipv4Address,
		VlanId:                m.VlanId,
		Customer:              m.Customer,
		CustomerName:          m.CustomerName,
		Type:                  m.Type,
		Parent:                m.Parent,
		Ipv6Prefix:            m.Ipv6Prefix,
		Ipv6Address:           m.Ipv6Address,
		DhcpServer:            m.DhcpServer,
		DhcpServerAddressPool: m.DhcpServerAddressPool,
		DhcpServerNetmask:     m.DhcpServerNetmask,
		DhcpServerDnsServers:  m.DhcpServerDnsServers,
		DhcpServerLeaseTime:   m.DhcpServerLeaseTime,
		Visible:               m.Visible,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}
}
