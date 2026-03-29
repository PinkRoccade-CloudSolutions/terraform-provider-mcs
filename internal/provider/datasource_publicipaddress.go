package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
)

var _ datasource.DataSource = &PublicIPAddressDataSource{}

type PublicIPAddressDataSource struct {
	client *apiclient.Client
}

type PublicIPAddressDataSourceModel struct {
	IPAddress         types.String               `tfsdk:"ip_address"`
	Id                types.String               `tfsdk:"id"`
	Pool              types.String               `tfsdk:"pool"`
	Description       types.String               `tfsdk:"description"`
	Status            types.String               `tfsdk:"status"`
	Type              types.String               `tfsdk:"type"`
	Customer          types.String               `tfsdk:"customer"`
	PublicIPAddresses []PublicIPAddressListModel  `tfsdk:"public_ip_addresses"`
}

type PublicIPAddressListModel struct {
	Id          types.String `tfsdk:"id"`
	IPAddress   types.String `tfsdk:"ip_address"`
	Pool        types.String `tfsdk:"pool"`
	Description types.String `tfsdk:"description"`
	Status      types.String `tfsdk:"status"`
	Type        types.String `tfsdk:"type"`
	Customer    types.String `tfsdk:"customer"`
}

type publicIPAddressDSAPIModel struct {
	Id          string  `json:"id"`
	IPAddress   string  `json:"ip_address"`
	Pool        *string `json:"pool"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
	Type        string  `json:"type"`
	Customer    *string `json:"customer"`
}

func NewPublicIPAddressDataSource() datasource.DataSource {
	return &PublicIPAddressDataSource{}
}

func (d *PublicIPAddressDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_ip_address"
}

func (d *PublicIPAddressDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	addrAttrs := map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true},
		"ip_address":  schema.StringAttribute{Computed: true},
		"pool":        schema.StringAttribute{Computed: true},
		"description": schema.StringAttribute{Computed: true},
		"status":      schema.StringAttribute{Computed: true},
		"type":        schema.StringAttribute{Computed: true},
		"customer":    schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS public IP addresses. Set `ip_address` or `id` to fetch a single address, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"ip_address": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Exact public IP address to look up, or the address of the matched entry.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific public IP address to look up.",
			},
			"pool": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the IP pool (set when a single address is matched).",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description (set when a single address is matched).",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Status: available, assigned, or reserved.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Type: nat, vip, or loadbalancer.",
			},
			"customer": schema.StringAttribute{
				Computed:    true,
				Description: "Customer identifier (set when a single address is matched).",
			},
			"public_ip_addresses": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All public IP addresses (populated when neither `ip_address` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: addrAttrs},
			},
		},
	}
}

func (d *PublicIPAddressDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PublicIPAddressDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PublicIPAddressDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var addr publicIPAddressDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/networking/publicipaddresss/%s/", config.Id.ValueString()), &addr)
		if err != nil {
			resp.Diagnostics.AddError("Error reading public IP address", err.Error())
			return
		}
		setSinglePublicIPAddress(&config, &addr)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	path := "/api/networking/publicipaddresss/?page_size=1000"
	if !config.IPAddress.IsNull() && config.IPAddress.ValueString() != "" {
		path += "&ip_address__icontains=" + url.QueryEscape(config.IPAddress.ValueString())
	}

	var page struct {
		Results []publicIPAddressDSAPIModel `json:"results"`
	}
	if err := d.client.Get(ctx, path, &page); err != nil {
		resp.Diagnostics.AddError("Error reading public IP addresses", err.Error())
		return
	}

	if !config.IPAddress.IsNull() && config.IPAddress.ValueString() != "" {
		var match *publicIPAddressDSAPIModel
		for i := range page.Results {
			if page.Results[i].IPAddress == config.IPAddress.ValueString() {
				match = &page.Results[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Public IP address not found",
				fmt.Sprintf("No public IP address with exact value %q was found.", config.IPAddress.ValueString()))
			return
		}
		setSinglePublicIPAddress(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := PublicIPAddressDataSourceModel{
		IPAddress:         types.StringNull(),
		Id:                types.StringNull(),
		Pool:              types.StringNull(),
		Description:       types.StringNull(),
		Status:            types.StringNull(),
		Type:              types.StringNull(),
		Customer:          types.StringNull(),
		PublicIPAddresses: make([]PublicIPAddressListModel, 0, len(page.Results)),
	}
	for _, item := range page.Results {
		state.PublicIPAddresses = append(state.PublicIPAddresses, toPublicIPAddressListModel(&item))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSinglePublicIPAddress(state *PublicIPAddressDataSourceModel, addr *publicIPAddressDSAPIModel) {
	state.Id = types.StringValue(addr.Id)
	state.IPAddress = types.StringValue(addr.IPAddress)
	state.Description = types.StringValue(addr.Description)
	state.Status = types.StringValue(addr.Status)
	state.Type = types.StringValue(addr.Type)
	if addr.Pool != nil {
		state.Pool = types.StringValue(*addr.Pool)
	} else {
		state.Pool = types.StringNull()
	}
	if addr.Customer != nil {
		state.Customer = types.StringValue(*addr.Customer)
	} else {
		state.Customer = types.StringNull()
	}
	state.PublicIPAddresses = []PublicIPAddressListModel{}
}

func toPublicIPAddressListModel(addr *publicIPAddressDSAPIModel) PublicIPAddressListModel {
	pool := types.StringNull()
	if addr.Pool != nil {
		pool = types.StringValue(*addr.Pool)
	}
	customer := types.StringNull()
	if addr.Customer != nil {
		customer = types.StringValue(*addr.Customer)
	}
	return PublicIPAddressListModel{
		Id:          types.StringValue(addr.Id),
		IPAddress:   types.StringValue(addr.IPAddress),
		Pool:        pool,
		Description: types.StringValue(addr.Description),
		Status:      types.StringValue(addr.Status),
		Type:        types.StringValue(addr.Type),
		Customer:    customer,
	}
}
