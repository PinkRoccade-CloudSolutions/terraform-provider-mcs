package provider

import (
	"context"
	"fmt"

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
	Networks []NetworkModel `tfsdk:"networks"`
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
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"networks": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
					},
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

func (d *NetworkDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var page struct {
		Results []networkAPIModel `json:"results"`
	}
	err := d.client.Get(ctx, "/api/networking/networks/?page_size=1000", &page)
	if err != nil {
		resp.Diagnostics.AddError("Error reading networks", err.Error())
		return
	}

	state := NetworkDataSourceModel{
		Networks: make([]NetworkModel, 0, len(page.Results)),
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
