package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/pinkroccade/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &DomainDataSource{}

type DomainDataSource struct {
	client *apiclient.Client
}

type DomainDataSourceModel struct {
	Domains []DomainModel `tfsdk:"domains"`
}

type DomainModel struct {
	Id          types.Int64  `tfsdk:"id"`
	Uuid        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Adom        types.String `tfsdk:"adom"`
	Zone        types.String `tfsdk:"zone"`
}

type domainAPIModel struct {
	Id          int    `json:"id"`
	Uuid        string `json:"uuid"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Adom        string `json:"adom"`
	Zone        string `json:"zone"`
}

func NewDomainDataSource() datasource.DataSource {
	return &DomainDataSource{}
}

func (d *DomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *DomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"domains": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
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
						"zone": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *DomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DomainDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var page struct {
		Results []domainAPIModel `json:"results"`
	}
	err := d.client.Get(ctx, "/api/tenant/domains/?page_size=1000", &page)
	if err != nil {
		resp.Diagnostics.AddError("Error reading domains", err.Error())
		return
	}

	state := DomainDataSourceModel{
		Domains: make([]DomainModel, 0, len(page.Results)),
	}
	for _, item := range page.Results {
		state.Domains = append(state.Domains, DomainModel{
			Id:          types.Int64Value(int64(item.Id)),
			Uuid:        types.StringValue(item.Uuid),
			Name:        types.StringValue(item.Name),
			Description: types.StringValue(item.Description),
			Adom:        types.StringValue(item.Adom),
			Zone:        types.StringValue(item.Zone),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
