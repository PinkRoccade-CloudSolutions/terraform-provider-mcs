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
	Name        types.String  `tfsdk:"name"`
	Id          types.Int64   `tfsdk:"id"`
	Uuid        types.String  `tfsdk:"uuid"`
	Description types.String  `tfsdk:"description"`
	Adom        types.String  `tfsdk:"adom"`
	Zone        types.String  `tfsdk:"zone"`
	Domains     []DomainModel `tfsdk:"domains"`
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
	domainAttrs := map[string]schema.Attribute{
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
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS domains. Set `name` to fetch a single domain by exact name, or omit it to list all domains.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact domain name to look up. When set, the data source returns a single domain and the `domains` list is empty.",
			},
			"id": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the matched domain (only set when `name` is provided).",
			},
			"uuid": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the matched domain (only set when `name` is provided).",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the matched domain (only set when `name` is provided).",
			},
			"adom": schema.StringAttribute{
				Computed:    true,
				Description: "ADOM of the matched domain (only set when `name` is provided).",
			},
			"zone": schema.StringAttribute{
				Computed:    true,
				Description: "Zone of the matched domain (only set when `name` is provided).",
			},
			"domains": schema.ListNestedAttribute{
				Computed:    true,
				Description: "All domains (populated when `name` is not set).",
				NestedObject: schema.NestedAttributeObject{
					Attributes: domainAttrs,
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

func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DomainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var page struct {
		Results []domainAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, "/api/tenant/domains/?page_size=1000", &page); err != nil {
		resp.Diagnostics.AddError("Error reading domains", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *domainAPIModel
		for i := range page.Results {
			if page.Results[i].Name == config.Name.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Domain not found",
				fmt.Sprintf("No domain with exact name %q was found.", config.Name.ValueString()))
			return
		}
		config.Id = types.Int64Value(int64(match.Id))
		config.Uuid = types.StringValue(match.Uuid)
		config.Description = types.StringValue(match.Description)
		config.Adom = types.StringValue(match.Adom)
		config.Zone = types.StringValue(match.Zone)
		config.Domains = []DomainModel{}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := DomainDataSourceModel{
		Name:        types.StringNull(),
		Id:          types.Int64Null(),
		Uuid:        types.StringNull(),
		Description: types.StringNull(),
		Adom:        types.StringNull(),
		Zone:        types.StringNull(),
		Domains:     make([]DomainModel, 0, len(page.Results)),
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
