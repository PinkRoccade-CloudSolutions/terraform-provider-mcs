package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DnsDomainDataSource{}

type DnsDomainDataSource struct {
	client *apiclient.Client
}

type DnsDomainDataSourceModel struct {
	Name    types.String       `tfsdk:"name"`
	Type    types.String       `tfsdk:"type"`
	Domains []DnsDomainListModel `tfsdk:"domains"`
}

type DnsDomainListModel struct {
	UUID         types.String `tfsdk:"uuid"`
	Name         types.String `tfsdk:"name"`
	Comment      types.String `tfsdk:"comment"`
	Enddate      types.String `tfsdk:"enddate"`
	Customer     types.String `tfsdk:"customer"`
	ProviderName types.String `tfsdk:"provider_name"`
	Type         types.String `tfsdk:"type"`
}

type dnsDomainAPIModel struct {
	UUID     string                  `json:"uuid"`
	Name     string                  `json:"name"`
	Comment  string                  `json:"comment"`
	Enddate  *string                 `json:"enddate"`
	Customer *string                 `json:"customer"`
	Provider integrationMinimalModel `json:"provider"`
	Type     string                  `json:"type"`
}

type integrationMinimalModel struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func NewDnsDomainDataSource() datasource.DataSource {
	return &DnsDomainDataSource{}
}

func (d *DnsDomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_domain"
}

func (d *DnsDomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	itemAttrs := map[string]schema.Attribute{
		"uuid":          schema.StringAttribute{Computed: true, Description: "UUID of the DNS domain."},
		"name":          schema.StringAttribute{Computed: true, Description: "Full zone name (e.g. domain.nl)."},
		"comment":       schema.StringAttribute{Computed: true, Description: "Comment for the domain."},
		"enddate":       schema.StringAttribute{Computed: true, Description: "End date for the domain, if known."},
		"customer":      schema.StringAttribute{Computed: true, Description: "Customer associated with the domain."},
		"provider_name": schema.StringAttribute{Computed: true, Description: "Name of the DNS provider integration."},
		"type":          schema.StringAttribute{Computed: true, Description: "Domain type: external or internal."},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS DNS domains, optionally filtered by name or type.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter domains by name (case-insensitive contains match).",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter domains by type: 'external' or 'internal'.",
			},
			"domains": schema.ListNestedAttribute{
				Computed:    true,
				Description: "List of matching DNS domains.",
				NestedObject: schema.NestedAttributeObject{
					Attributes: itemAttrs,
				},
			},
		},
	}
}

func (d *DnsDomainDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DnsDomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DnsDomainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	path := "/api/dns/domains/"
	params := url.Values{}
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		params.Set("name__icontains", config.Name.ValueString())
	}
	if !config.Type.IsNull() && config.Type.ValueString() != "" {
		params.Set("type", config.Type.ValueString())
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	raw, err := d.client.ListAll(ctx, path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading DNS domains", err.Error())
		return
	}

	state := DnsDomainDataSourceModel{
		Name:    config.Name,
		Type:    config.Type,
		Domains: make([]DnsDomainListModel, 0, len(raw)),
	}

	for _, item := range raw {
		var domain dnsDomainAPIModel
		if err := json.Unmarshal(item, &domain); err != nil {
			resp.Diagnostics.AddError("Error parsing DNS domain", err.Error())
			return
		}
		state.Domains = append(state.Domains, dnsDomainToListModel(&domain))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func dnsDomainToListModel(item *dnsDomainAPIModel) DnsDomainListModel {
	m := DnsDomainListModel{
		UUID:         types.StringValue(item.UUID),
		Name:         types.StringValue(item.Name),
		Comment:      types.StringValue(item.Comment),
		ProviderName: types.StringValue(item.Provider.Name),
		Type:         types.StringValue(item.Type),
	}
	if item.Enddate != nil {
		m.Enddate = types.StringValue(*item.Enddate)
	} else {
		m.Enddate = types.StringNull()
	}
	if item.Customer != nil {
		m.Customer = types.StringValue(*item.Customer)
	} else {
		m.Customer = types.StringNull()
	}
	return m
}
