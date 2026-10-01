package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NetworkProductsDataSource{}

type NetworkProductsDataSource struct {
	client *apiclient.Client
}

type NetworkProductsDataSourceModel struct {
	Names types.List `tfsdk:"names"`
}

type networkProductAPIModel struct {
	Name string `json:"name"`
}

func NewNetworkProductsDataSource() datasource.DataSource {
	return &NetworkProductsDataSource{}
}

func (d *NetworkProductsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_products"
}

func (d *NetworkProductsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists the products a network can be created under (the `product` argument of `mcs_network`). " +
			"Each product owns a PRODID band.",
		Attributes: map[string]schema.Attribute{
			"names": schema.ListAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Names of the available network products.",
			},
		},
	}
}

func (d *NetworkProductsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetworkProductsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	// The endpoint answers with a plain array, not a paginated envelope.
	var products []networkProductAPIModel
	if err := d.client.Get(ctx, "/api/v3/networking/network-products/", &products); err != nil {
		resp.Diagnostics.AddError("Error reading network products", err.Error())
		return
	}

	names := make([]string, 0, len(products))
	for _, p := range products {
		names = append(names, p.Name)
	}
	state := NetworkProductsDataSourceModel{
		Names: computedListValue(ctx, types.StringType, names, &resp.Diagnostics),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
