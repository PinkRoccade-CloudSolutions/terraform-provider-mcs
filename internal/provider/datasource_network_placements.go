package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NetworkPlacementsDataSource{}

type NetworkPlacementsDataSource struct {
	client *apiclient.Client
}

type NetworkPlacementsDataSourceModel struct {
	DomainUuid               types.String `tfsdk:"domain_uuid"`
	Standard                 types.List   `tfsdk:"standard"`
	Kubernetes               types.List   `tfsdk:"kubernetes"`
	KubernetesEqualsStandard types.Bool   `tfsdk:"kubernetes_equals_standard"`
}

type networkPlacementsAPIModel struct {
	Standard                 []datacenterRefAPIModel `json:"standard"`
	Kubernetes               []datacenterRefAPIModel `json:"kubernetes"`
	KubernetesEqualsStandard bool                    `json:"kubernetes_equals_standard"`
}

func NewNetworkPlacementsDataSource() datasource.DataSource {
	return &NetworkPlacementsDataSource{}
}

func (d *NetworkPlacementsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_placements"
}

func (d *NetworkPlacementsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	datacenters := func(desc string) schema.ListNestedAttribute {
		return schema.ListNestedAttribute{
			Computed:    true,
			Description: desc,
			NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"id":   schema.Int64Attribute{Computed: true, Description: "Datacenter ID."},
				"name": schema.StringAttribute{Computed: true, Description: "Datacenter name."},
			}},
		}
	}
	resp.Schema = schema.Schema{
		Description: "Look up the datacenters a V2 network deploy to a domain would go to, per `placement` of `mcs_network`.",
		Attributes: map[string]schema.Attribute{
			"domain_uuid": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the domain (VDOM), e.g. `data.mcs_domain.x.uuid`.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"standard": datacenters("Datacenters for `placement = \"standard\"`. Empty means the zone is not ready for V2."),
			"kubernetes": datacenters("Datacenters for `placement = \"kubernetes\"`: the standard ones plus the zone's " +
				"additional datacenters."),
			"kubernetes_equals_standard": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether both placements deploy to the same datacenters.",
			},
		},
	}
}

func (d *NetworkPlacementsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetworkPlacementsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config NetworkPlacementsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var placements networkPlacementsAPIModel
	path := "/api/v3/networking/network-placements/?" + url.Values{"domain": {config.DomainUuid.ValueString()}}.Encode()
	if err := d.client.Get(ctx, path, &placements); err != nil {
		resp.Diagnostics.AddError("Error reading network placements", err.Error())
		return
	}

	config.Standard = datacenterRefList(placements.Standard)
	config.Kubernetes = datacenterRefList(placements.Kubernetes)
	config.KubernetesEqualsStandard = types.BoolValue(placements.KubernetesEqualsStandard)
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
