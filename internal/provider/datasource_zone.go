package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &ZoneDataSource{}

type ZoneDataSource struct {
	client *apiclient.Client
}

type ZoneDataSourceModel struct {
	Name             types.String            `tfsdk:"name"`
	Id               types.Int64             `tfsdk:"id"`
	Uuid             types.String            `tfsdk:"uuid"`
	Description      types.String            `tfsdk:"description"`
	Adom             types.String            `tfsdk:"adom"`
	TransitVrf       types.String            `tfsdk:"transit_vrf"`
	Loadbalancers    []ZoneLoadbalancerModel `tfsdk:"loadbalancers"`
	EdgeRootNetworks types.List              `tfsdk:"edge_root_networks"`
	Zones            []ZoneModel             `tfsdk:"zones"`
}

type ZoneModel struct {
	Id               types.Int64             `tfsdk:"id"`
	Uuid             types.String            `tfsdk:"uuid"`
	Name             types.String            `tfsdk:"name"`
	Description      types.String            `tfsdk:"description"`
	Adom             types.String            `tfsdk:"adom"`
	TransitVrf       types.String            `tfsdk:"transit_vrf"`
	Loadbalancers    []ZoneLoadbalancerModel `tfsdk:"loadbalancers"`
	EdgeRootNetworks types.List              `tfsdk:"edge_root_networks"`
}

type ZoneLoadbalancerModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type zoneAPIModel struct {
	Id               int64                      `json:"id"`
	Uuid             string                     `json:"uuid"`
	Name             string                     `json:"name"`
	Description      string                     `json:"description"`
	Adom             string                     `json:"adom"`
	TransitVrf       string                     `json:"transit_vrf"`
	Loadbalancers    []zoneLoadbalancerAPIModel `json:"loadbalancers"`
	EdgeRootNetworks []string                   `json:"edge_root_networks"`
}

// zoneLoadbalancerAPIModel is the spec's MinimalDevice; its display name is `alias`.
type zoneLoadbalancerAPIModel struct {
	Id    string `json:"id"`
	Alias string `json:"alias"`
}

func NewZoneDataSource() datasource.DataSource {
	return &ZoneDataSource{}
}

func (d *ZoneDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_zone"
}

func (d *ZoneDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	loadbalancerAttrs := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Computed: true,
		},
		"name": schema.StringAttribute{
			Computed:    true,
			Description: "Alias of the load balancer device.",
		},
	}

	zoneAttrs := map[string]schema.Attribute{
		"id": schema.Int64Attribute{
			Computed: true,
		},
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
		"loadbalancers": schema.ListNestedAttribute{
			Computed:    true,
			Description: "Load balancers available in this zone.",
			NestedObject: schema.NestedAttributeObject{
				Attributes: loadbalancerAttrs,
			},
		},
		"edge_root_networks": schema.ListAttribute{
			Computed:    true,
			ElementType: types.StringType,
			Description: "UUIDs of the edge root networks of this zone.",
		},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS zones. Set `name` to fetch a single zone by exact name, or omit it to list all zones.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact zone name to look up. When set, the data source returns a single zone and the `zones` list is empty.",
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "Numeric ID of the matched zone (only set when `name` is provided).",
			},
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the matched zone (only set when `name` is provided).",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the matched zone (only set when `name` is provided).",
			},
			"adom": schema.StringAttribute{
				Computed:    true,
				Description: "ADOM of the matched zone (only set when `name` is provided).",
			},
			"transit_vrf": schema.StringAttribute{
				Computed:    true,
				Description: "Transit VRF of the matched zone (only set when `name` is provided).",
			},
			"loadbalancers": schema.ListNestedAttribute{
				Computed:    true,
				Description: "Load balancers available in the matched zone (only set when `name` is provided).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: loadbalancerAttrs,
				},
			},
			"edge_root_networks": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "UUIDs of the edge root networks of the matched zone (only set when `name` is provided).",
			},
			"zones": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All zones (populated when `name` is not set).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: zoneAttrs,
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

func (d *ZoneDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ZoneDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/api/networking/zones/"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "?name__icontains=" + url.QueryEscape(config.Name.ValueString())
	}

	items, err := listAll[zoneAPIModel](ctx, d.client, path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading zones", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *zoneAPIModel
		for i := range items {
			if items[i].Name == config.Name.ValueString() {
				match = &items[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Zone not found",
				fmt.Sprintf("No zone with exact name %q was found.", config.Name.ValueString()))
			return
		}
		m := toZoneModel(ctx, match, &resp.Diagnostics)
		config.Id = m.Id
		config.Uuid = m.Uuid
		config.Description = m.Description
		config.Adom = m.Adom
		config.TransitVrf = m.TransitVrf
		config.Loadbalancers = m.Loadbalancers
		config.EdgeRootNetworks = m.EdgeRootNetworks
		config.Zones = []ZoneModel{}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := ZoneDataSourceModel{
		Name:             types.StringNull(),
		Id:               types.Int64Null(),
		Uuid:             types.StringNull(),
		Description:      types.StringNull(),
		Adom:             types.StringNull(),
		TransitVrf:       types.StringNull(),
		Loadbalancers:    []ZoneLoadbalancerModel{},
		EdgeRootNetworks: types.ListNull(types.StringType),
		Zones:            make([]ZoneModel, 0, len(items)),
	}
	for i := range items {
		state.Zones = append(state.Zones, toZoneModel(ctx, &items[i], &resp.Diagnostics))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func toZoneModel(ctx context.Context, z *zoneAPIModel, diags *diag.Diagnostics) ZoneModel {
	return ZoneModel{
		Id:               types.Int64Value(z.Id),
		Uuid:             types.StringValue(z.Uuid),
		Name:             types.StringValue(z.Name),
		Description:      types.StringValue(z.Description),
		Adom:             types.StringValue(z.Adom),
		TransitVrf:       types.StringValue(z.TransitVrf),
		Loadbalancers:    mapLoadbalancers(z.Loadbalancers),
		EdgeRootNetworks: computedListValue(ctx, types.StringType, z.EdgeRootNetworks, diags),
	}
}

func mapLoadbalancers(lbs []zoneLoadbalancerAPIModel) []ZoneLoadbalancerModel {
	out := make([]ZoneLoadbalancerModel, 0, len(lbs))
	for _, lb := range lbs {
		out = append(out, ZoneLoadbalancerModel{
			Id:   types.StringValue(lb.Id),
			Name: types.StringValue(lb.Alias),
		})
	}
	return out
}
