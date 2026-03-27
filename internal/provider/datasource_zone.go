package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &ZoneDataSource{}

type ZoneDataSource struct {
	client *apiclient.Client
}

type ZoneDataSourceModel struct {
	Zones []ZoneModel `tfsdk:"zones"`
}

type ZoneModel struct {
	Uuid        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Adom        types.String `tfsdk:"adom"`
	TransitVrf  types.String `tfsdk:"transit_vrf"`
}

type zoneAPIModel struct {
	Uuid        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Adom        string `json:"adom"`
	TransitVrf  string `json:"transit_vrf"`
}

func NewZoneDataSource() datasource.DataSource {
	return &ZoneDataSource{}
}

func (d *ZoneDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *ZoneDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"zones": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"uuid": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"adom": schema.StringAttribute{
							Computed: true,
						},
						"transit_vrf": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *ZoneDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ZoneDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var page struct {
		Results []zoneAPIModel `json:"results"`
	}
	err := d.client.Get(ctx, "/api/networking/zones/?page_size=1000", &page)
	if err != nil {
		resp.Diagnostics.AddError("Error reading zones", err.Error())
		return
	}

	state := ZoneDataSourceModel{
		Zones: make([]ZoneModel, 0, len(page.Results)),
	}
	for _, item := range page.Results {
		state.Zones = append(state.Zones, ZoneModel{
			Uuid:        types.StringValue(item.Uuid),
			Name:        types.StringValue(item.Name),
			Description: types.StringValue(item.Description),
			Adom:        types.StringValue(item.Adom),
			TransitVrf:  types.StringValue(item.TransitVrf),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
