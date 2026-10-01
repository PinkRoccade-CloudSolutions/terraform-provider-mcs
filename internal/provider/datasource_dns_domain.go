package provider

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &DnsDomainDataSource{}

type DnsDomainDataSource struct {
	client *apiclient.Client
}

type DnsDomainDataSourceModel struct {
	Name     types.String         `tfsdk:"name"`
	Type     types.String         `tfsdk:"type"`
	Customer types.String         `tfsdk:"customer"`
	Provider types.Int64          `tfsdk:"provider_id"`
	ZoneType types.String         `tfsdk:"zone_type"`
	Domains  []DnsDomainListModel `tfsdk:"domains"`
}

type DnsDomainListModel struct {
	UUID         types.String `tfsdk:"uuid"`
	Name         types.String `tfsdk:"name"`
	Comment      types.String `tfsdk:"comment"`
	Enddate      types.String `tfsdk:"enddate"`
	Customer     types.String `tfsdk:"customer"`
	ProviderID   types.Int64  `tfsdk:"provider_id"`
	ProviderName types.String `tfsdk:"provider_name"`
	Type         types.String `tfsdk:"type"`
	ZoneType     types.String `tfsdk:"zone_type"`
}

type dnsDomainAPIModel struct {
	UUID     string                  `json:"uuid,omitempty"`
	Name     string                  `json:"name"`
	Comment  string                  `json:"comment"`
	Enddate  *string                 `json:"enddate"`
	Customer *string                 `json:"customer"`
	Provider integrationMinimalModel `json:"provider"`
	Type     string                  `json:"type"`
	ZoneType string                  `json:"zone_type"`
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
		"provider_id":   schema.Int64Attribute{Computed: true, Description: "ID of the DNS provider integration."},
		"provider_name": schema.StringAttribute{Computed: true, Description: "Name of the DNS provider integration."},
		"type":          schema.StringAttribute{Computed: true, Description: "Domain type: external or internal."},
		"zone_type":     schema.StringAttribute{Computed: true, Description: "Zone type: forward or reverse."},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS DNS domains, optionally filtered by name, type, zone type, customer or provider.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Filter domains by name (case-insensitive contains match).",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter domains by type: 'external' or 'internal'.",
			},
			"customer": schema.StringAttribute{
				Optional:    true,
				Description: "Filter domains by customer.",
			},
			"provider_id": schema.Int64Attribute{
				Optional:    true,
				Description: "Filter domains by DNS provider integration ID.",
			},
			"zone_type": schema.StringAttribute{
				Optional:    true,
				Description: "Filter domains by zone type: 'forward' or 'reverse'.",
				Validators:  []validator.String{stringvalidator.OneOf("forward", "reverse")},
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
	if !config.Customer.IsNull() && config.Customer.ValueString() != "" {
		params.Set("customer", config.Customer.ValueString())
	}
	if !config.Provider.IsNull() {
		params.Set("provider", strconv.FormatInt(config.Provider.ValueInt64(), 10))
	}
	if !config.ZoneType.IsNull() && config.ZoneType.ValueString() != "" {
		params.Set("zone_type", config.ZoneType.ValueString())
	}
	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	items, err := listAll[dnsDomainAPIModel](ctx, d.client, path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading DNS domains", err.Error())
		return
	}

	state := DnsDomainDataSourceModel{
		Name:     config.Name,
		Type:     config.Type,
		Customer: config.Customer,
		Provider: config.Provider,
		ZoneType: config.ZoneType,
		Domains:  make([]DnsDomainListModel, 0, len(items)),
	}
	for i := range items {
		state.Domains = append(state.Domains, dnsDomainToListModel(&items[i]))
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func dnsDomainToListModel(item *dnsDomainAPIModel) DnsDomainListModel {
	return DnsDomainListModel{
		UUID:         types.StringValue(item.UUID),
		Name:         types.StringValue(item.Name),
		Comment:      types.StringValue(item.Comment),
		Enddate:      types.StringPointerValue(item.Enddate),
		Customer:     types.StringPointerValue(item.Customer),
		ProviderID:   types.Int64Value(int64(item.Provider.ID)),
		ProviderName: types.StringValue(item.Provider.Name),
		Type:         types.StringValue(item.Type),
		ZoneType:     types.StringValue(item.ZoneType),
	}
}
