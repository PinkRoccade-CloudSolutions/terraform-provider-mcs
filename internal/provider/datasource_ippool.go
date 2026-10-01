package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &IPPoolDataSource{}

type IPPoolDataSource struct {
	client *apiclient.Client
}

type IPPoolDataSourceModel struct {
	Name     types.String  `tfsdk:"name"`
	Id       types.String  `tfsdk:"id"`
	Subnet   types.String  `tfsdk:"subnet"`
	Customer types.String  `tfsdk:"customer"`
	Type     types.String  `tfsdk:"type"`
	TotalIPs types.String  `tfsdk:"total_ips"`
	FreeIPs  types.String  `tfsdk:"free_ips"`
	IPPools  []IPPoolModel `tfsdk:"ip_pools"`
}

type IPPoolModel struct {
	Id       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Subnet   types.String `tfsdk:"subnet"`
	Customer types.String `tfsdk:"customer"`
	Type     types.String `tfsdk:"type"`
	TotalIPs types.String `tfsdk:"total_ips"`
	FreeIPs  types.String `tfsdk:"free_ips"`
}

type ippoolAPIModel struct {
	Id       string      `json:"id"`
	Name     string      `json:"name"`
	Subnet   string      `json:"subnet"`
	Type     string      `json:"type"`
	Customer *string     `json:"customer"`
	TotalIPs ippoolCount `json:"total_ips"`
	FreeIPs  ippoolCount `json:"free_ips"`
}

// ippoolCount holds total_ips/free_ips, which the spec types as string; a JSON number is
// accepted as well.
type ippoolCount string

func (c *ippoolCount) UnmarshalJSON(b []byte) error {
	b = bytes.TrimSpace(b)
	if bytes.Equal(b, []byte("null")) {
		*c = ""
		return nil
	}
	if len(b) > 0 && b[0] == '"' {
		var s string
		if err := json.Unmarshal(b, &s); err != nil {
			return err
		}
		*c = ippoolCount(s)
		return nil
	}
	var n json.Number
	if err := json.Unmarshal(b, &n); err != nil {
		return fmt.Errorf("expected string or number: %w", err)
	}
	*c = ippoolCount(n.String())
	return nil
}

func NewIPPoolDataSource() datasource.DataSource {
	return &IPPoolDataSource{}
}

func (d *IPPoolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ippool"
}

func (d *IPPoolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	poolAttrs := map[string]schema.Attribute{
		"id":        schema.StringAttribute{Computed: true},
		"name":      schema.StringAttribute{Computed: true},
		"subnet":    schema.StringAttribute{Computed: true},
		"customer":  schema.StringAttribute{Computed: true},
		"type":      schema.StringAttribute{Computed: true},
		"total_ips": schema.StringAttribute{Computed: true},
		"free_ips":  schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS IP pools. Set `name` or `id` to fetch a single pool, or omit both to list all (optionally filtered by `customer` and `type`).",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact IP pool name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific IP pool to look up.",
			},
			"subnet": schema.StringAttribute{
				Computed:    true,
				Description: "Subnet of the matched IP pool.",
			},
			"customer": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Filter by customer; set to the customer of the matched IP pool.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Filter by pool type (nat, vip, loadbalancer); set to the type of the matched IP pool.",
			},
			"total_ips": schema.StringAttribute{
				Computed:    true,
				Description: "Total number of IP addresses in the matched pool.",
			},
			"free_ips": schema.StringAttribute{
				Computed:    true,
				Description: "Number of free IP addresses in the matched pool.",
			},
			"ip_pools": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All IP pools (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: poolAttrs},
			},
		},
	}
}

func (d *IPPoolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *IPPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config IPPoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var pool ippoolAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/networking/ippools/%s/", config.Id.ValueString()), &pool)
		if err != nil {
			resp.Diagnostics.AddError("Error reading IP pool", err.Error())
			return
		}
		setSingleIPPool(&config, &pool)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	q := url.Values{}
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		q.Set("name__icontains", config.Name.ValueString())
	}
	if !config.Customer.IsNull() && config.Customer.ValueString() != "" {
		q.Set("customer", config.Customer.ValueString())
	}
	if !config.Type.IsNull() && config.Type.ValueString() != "" {
		q.Set("type", config.Type.ValueString())
	}
	path := "/api/networking/ippools/"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}

	items, err := listAll[ippoolAPIModel](ctx, d.client, path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IP pools", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *ippoolAPIModel
		for i := range items {
			if items[i].Name == config.Name.ValueString() {
				match = &items[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("IP pool not found",
				fmt.Sprintf("No IP pool with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleIPPool(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := IPPoolDataSourceModel{
		Name:     types.StringNull(),
		Id:       types.StringNull(),
		Subnet:   types.StringNull(),
		Customer: config.Customer,
		Type:     config.Type,
		TotalIPs: types.StringNull(),
		FreeIPs:  types.StringNull(),
		IPPools:  make([]IPPoolModel, 0, len(items)),
	}
	if state.Customer.IsUnknown() {
		state.Customer = types.StringNull()
	}
	if state.Type.IsUnknown() {
		state.Type = types.StringNull()
	}
	for _, item := range items {
		state.IPPools = append(state.IPPools, IPPoolModel{
			Id:       types.StringValue(item.Id),
			Name:     types.StringValue(item.Name),
			Subnet:   types.StringValue(item.Subnet),
			Customer: types.StringPointerValue(item.Customer),
			Type:     types.StringValue(item.Type),
			TotalIPs: types.StringValue(string(item.TotalIPs)),
			FreeIPs:  types.StringValue(string(item.FreeIPs)),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleIPPool(state *IPPoolDataSourceModel, pool *ippoolAPIModel) {
	state.Id = types.StringValue(pool.Id)
	state.Name = types.StringValue(pool.Name)
	state.Subnet = types.StringValue(pool.Subnet)
	state.Customer = types.StringPointerValue(pool.Customer)
	state.Type = types.StringValue(pool.Type)
	state.TotalIPs = types.StringValue(string(pool.TotalIPs))
	state.FreeIPs = types.StringValue(string(pool.FreeIPs))
	state.IPPools = []IPPoolModel{}
}
