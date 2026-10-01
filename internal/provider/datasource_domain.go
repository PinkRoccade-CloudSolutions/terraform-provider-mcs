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

var _ datasource.DataSource = &DomainDataSource{}

type DomainDataSource struct {
	client *apiclient.Client
}

type DomainDataSourceModel struct {
	Name         types.String  `tfsdk:"name"`
	Id           types.Int64   `tfsdk:"id"`
	Uuid         types.String  `tfsdk:"uuid"`
	VdomId       types.String  `tfsdk:"vdom_id"`
	Description  types.String  `tfsdk:"description"`
	Adom         types.String  `tfsdk:"adom"`
	ZoneId       types.Int64   `tfsdk:"zone_id"`
	ZoneName     types.String  `tfsdk:"zone_name"`
	Customer     types.String  `tfsdk:"customer"`
	CustomerName types.String  `tfsdk:"customer_name"`
	Tag          types.List    `tfsdk:"tag"`
	CompErrors   types.Int64   `tfsdk:"comp_errors"`
	MaxNetworks  types.Int64   `tfsdk:"max_networks"`
	Domains      []DomainModel `tfsdk:"domains"`
}

type DomainModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Uuid         types.String `tfsdk:"uuid"`
	VdomId       types.String `tfsdk:"vdom_id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	Adom         types.String `tfsdk:"adom"`
	ZoneId       types.Int64  `tfsdk:"zone_id"`
	ZoneName     types.String `tfsdk:"zone_name"`
	Customer     types.String `tfsdk:"customer"`
	CustomerName types.String `tfsdk:"customer_name"`
	Tag          types.List   `tfsdk:"tag"`
	CompErrors   types.Int64  `tfsdk:"comp_errors"`
	MaxNetworks  types.Int64  `tfsdk:"max_networks"`
}

type domainZoneRefAPI struct {
	Id   *int64  `json:"id"`
	Name *string `json:"name"`
}

type domainCustomerAPI struct {
	Id         string `json:"id"`
	Contractid string `json:"contractid"`
	Name       string `json:"name"`
}

type domainAPIModel struct {
	Id          int64              `json:"id"`
	Uuid        string             `json:"uuid"`
	VdomId      string             `json:"vdom_id"`
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Adom        string             `json:"adom"`
	Zone        *domainZoneRefAPI  `json:"zone"`
	Customer    *domainCustomerAPI `json:"customer"`
	Tag         []int64            `json:"tag"`
	CompErrors  int64              `json:"comp_errors"`
	MaxNetworks int64              `json:"max_networks"`
}

func NewDomainDataSource() datasource.DataSource {
	return &DomainDataSource{}
}

func (d *DomainDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *DomainDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	domainAttrs := map[string]schema.Attribute{
		"id":            schema.Int64Attribute{Computed: true},
		"uuid":          schema.StringAttribute{Computed: true},
		"vdom_id":       schema.StringAttribute{Computed: true},
		"name":          schema.StringAttribute{Computed: true},
		"description":   schema.StringAttribute{Computed: true},
		"adom":          schema.StringAttribute{Computed: true},
		"zone_id":       schema.Int64Attribute{Computed: true},
		"zone_name":     schema.StringAttribute{Computed: true},
		"customer":      schema.StringAttribute{Computed: true},
		"customer_name": schema.StringAttribute{Computed: true},
		"tag":           schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
		"comp_errors":   schema.Int64Attribute{Computed: true},
		"max_networks":  schema.Int64Attribute{Computed: true},
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
			"vdom_id": schema.StringAttribute{
				Computed:    true,
				Description: "VDOM ID of the matched domain (only set when `name` is provided).",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the matched domain (only set when `name` is provided).",
			},
			"adom": schema.StringAttribute{
				Computed:    true,
				Description: "ADOM of the matched domain (only set when `name` is provided).",
			},
			"zone_id": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the network zone of the matched domain (only set when `name` is provided).",
			},
			"zone_name": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the network zone of the matched domain (only set when `name` is provided).",
			},
			"customer": schema.StringAttribute{
				Computed:    true,
				Description: "Customer ID of the matched domain (only set when `name` is provided).",
			},
			"customer_name": schema.StringAttribute{
				Computed:    true,
				Description: "Customer name of the matched domain (only set when `name` is provided).",
			},
			"tag": schema.ListAttribute{
				Computed:    true,
				ElementType: types.Int64Type,
				Description: "Tag IDs of the matched domain (only set when `name` is provided).",
			},
			"comp_errors": schema.Int64Attribute{
				Computed:    true,
				Description: "Number of compliance errors of the matched domain (only set when `name` is provided).",
			},
			"max_networks": schema.Int64Attribute{
				Computed:    true,
				Description: "Maximum number of networks of the matched domain (only set when `name` is provided).",
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

func domainModelFromAPI(ctx context.Context, item *domainAPIModel, diags *diag.Diagnostics) DomainModel {
	m := DomainModel{
		Id:           types.Int64Value(item.Id),
		Uuid:         types.StringValue(item.Uuid),
		VdomId:       types.StringValue(item.VdomId),
		Name:         types.StringValue(item.Name),
		Description:  types.StringValue(item.Description),
		Adom:         types.StringValue(item.Adom),
		ZoneId:       types.Int64Null(),
		ZoneName:     types.StringNull(),
		Customer:     types.StringNull(),
		CustomerName: types.StringNull(),
		Tag:          computedListValue(ctx, types.Int64Type, item.Tag, diags),
		CompErrors:   types.Int64Value(item.CompErrors),
		MaxNetworks:  types.Int64Value(item.MaxNetworks),
	}
	if item.Zone != nil {
		m.ZoneId = types.Int64PointerValue(item.Zone.Id)
		m.ZoneName = types.StringPointerValue(item.Zone.Name)
	}
	if item.Customer != nil {
		m.Customer = types.StringValue(item.Customer.Id)
		m.CustomerName = types.StringValue(item.Customer.Name)
	}
	return m
}

func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DomainDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := config.Name.ValueString()
	path := "/api/tenant/domains/"
	if name != "" {
		path += "?name__icontains=" + url.QueryEscape(name)
	}

	items, err := listAll[domainAPIModel](ctx, d.client, path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading domains", err.Error())
		return
	}

	if name != "" {
		var match *domainAPIModel
		for i := range items {
			if items[i].Name == name {
				match = &items[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Domain not found",
				fmt.Sprintf("No domain with exact name %q was found.", name))
			return
		}
		m := domainModelFromAPI(ctx, match, &resp.Diagnostics)
		config.Id = m.Id
		config.Uuid = m.Uuid
		config.VdomId = m.VdomId
		config.Description = m.Description
		config.Adom = m.Adom
		config.ZoneId = m.ZoneId
		config.ZoneName = m.ZoneName
		config.Customer = m.Customer
		config.CustomerName = m.CustomerName
		config.Tag = m.Tag
		config.CompErrors = m.CompErrors
		config.MaxNetworks = m.MaxNetworks
		config.Domains = []DomainModel{}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := DomainDataSourceModel{
		Name:         types.StringNull(),
		Id:           types.Int64Null(),
		Uuid:         types.StringNull(),
		VdomId:       types.StringNull(),
		Description:  types.StringNull(),
		Adom:         types.StringNull(),
		ZoneId:       types.Int64Null(),
		ZoneName:     types.StringNull(),
		Customer:     types.StringNull(),
		CustomerName: types.StringNull(),
		Tag:          types.ListNull(types.Int64Type),
		CompErrors:   types.Int64Null(),
		MaxNetworks:  types.Int64Null(),
		Domains:      make([]DomainModel, 0, len(items)),
	}
	for i := range items {
		state.Domains = append(state.Domains, domainModelFromAPI(ctx, &items[i], &resp.Diagnostics))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
