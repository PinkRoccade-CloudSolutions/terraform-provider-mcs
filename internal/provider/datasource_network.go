package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &NetworkDataSource{}

type NetworkDataSource struct {
	client *apiclient.Client
}

type NetworkDataSourceModel struct {
	Name       types.String   `tfsdk:"name"`
	Id         types.String   `tfsdk:"id"`
	Ipv4Prefix types.String   `tfsdk:"ipv4_prefix"`
	VlanId     types.Int64    `tfsdk:"vlan_id"`
	Networks   []NetworkModel `tfsdk:"networks"`
}

type NetworkModel struct {
	Id         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Ipv4Prefix types.String `tfsdk:"ipv4_prefix"`
	VlanId     types.Int64  `tfsdk:"vlan_id"`
}

type networkAPIModel struct {
	Id         string `json:"id"`
	Name       string `json:"name"`
	Ipv4Prefix string `json:"ipv4_prefix"`
	VlanId     int64  `json:"vlan_id"`
}

func NewNetworkDataSource() datasource.DataSource {
	return &NetworkDataSource{}
}

func (d *NetworkDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network"
}

func (d *NetworkDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	networkAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed: true,
		},
		"ipv4_prefix": schema.StringAttribute{
			Computed: true,
		},
		"vlan_id": schema.Int64Attribute{
			Computed: true,
		},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS networks. Set `name` to fetch a single network by exact name, or omit it to list all networks.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact network name to look up. When set, the data source returns a single network and the `networks` list is empty.",
			},
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "ID of the matched network (only set when `name` is provided).",
			},
			"ipv4_prefix": schema.StringAttribute{
				Computed:    true,
				Description: "IPv4 prefix of the matched network (only set when `name` is provided).",
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

	path := "/api/networking/networks/?page_size=1000"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "&name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	var page struct {
		Results []networkAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading networks", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *networkAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Network not found",
				fmt.Sprintf("No network with exact name %q was found.", config.Name.ValueString()))
			return
		}
		config.Id = types.StringValue(match.Id)
		config.Ipv4Prefix = types.StringValue(match.Ipv4Prefix)
		config.VlanId = types.Int64Value(match.VlanId)
		config.Networks = []NetworkModel{}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := NetworkDataSourceModel{
		Name:       types.StringNull(),
		Id:         types.StringNull(),
		Ipv4Prefix: types.StringNull(),
		VlanId:     types.Int64Null(),
		Networks:   make([]NetworkModel, 0, len(page.Results)),
	}
	for _, item := range page.Results {
		state.Networks = append(state.Networks, NetworkModel{
			Id:         types.StringValue(item.Id),
			Name:       types.StringValue(item.Name),
			Ipv4Prefix: types.StringValue(item.Ipv4Prefix),
			VlanId:     types.Int64Value(item.VlanId),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
