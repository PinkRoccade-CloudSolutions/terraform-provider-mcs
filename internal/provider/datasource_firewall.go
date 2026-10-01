package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &FirewallDataSource{}

type FirewallDataSource struct {
	client *apiclient.Client
}

type FirewallDataSourceModel struct {
	FirewallModel
	Firewalls []FirewallModel `tfsdk:"firewalls"`
}

type FirewallModel struct {
	Id                       types.String `tfsdk:"id"`
	Name                     types.String `tfsdk:"name"`
	Description              types.String `tfsdk:"description"`
	Customer                 types.String `tfsdk:"customer"`
	CustomerName             types.String `tfsdk:"customer_name"`
	Type                     types.String `tfsdk:"type"`
	Device                   types.String `tfsdk:"device"`
	DeviceName               types.String `tfsdk:"device_name"`
	Platform                 types.String `tfsdk:"platform"`
	SupportsThreatProtection types.Bool   `tfsdk:"supports_threat_protection"`
	Context                  types.String `tfsdk:"context"`
	ExternalInterface        types.String `tfsdk:"external_interface"`
	InternalInterface        types.String `tfsdk:"internal_interface"`
	DefaultLogProfile        types.String `tfsdk:"default_log_profile"`
	DefaultProtectProfile    types.String `tfsdk:"default_protect_profile"`
	MultiTenant              types.Bool   `tfsdk:"multi_tenant"`
	TagName                  types.String `tfsdk:"tag_name"`
	NatIpSyncEnabled         types.Bool   `tfsdk:"nat_ip_sync_enabled"`
}

// firewallAPIModel is the Firewall schema of /api/networking/firewalls/, shared with the mcs_firewall resource.
type firewallAPIModel struct {
	Id                       string `json:"id"`
	Name                     string `json:"name"`
	Description              string `json:"description"`
	Customer                 string `json:"customer"`
	CustomerName             string `json:"customer_name"`
	Type                     string `json:"type"`
	Device                   string `json:"device"`
	DeviceName               string `json:"device_name"`
	Platform                 string `json:"platform"`
	SupportsThreatProtection bool   `json:"supports_threat_protection"`
	Context                  string `json:"context"`
	ExternalInterface        string `json:"external_interface"`
	InternalInterface        string `json:"internal_interface"`
	DefaultLogProfile        string `json:"default_log_profile"`
	DefaultProtectProfile    string `json:"default_protect_profile"`
	MultiTenant              bool   `json:"multi_tenant"`
	TagName                  string `json:"tag_name"`
	NatIpSyncEnabled         bool   `json:"nat_ip_sync_enabled"`
}

func NewFirewallDataSource() datasource.DataSource {
	return &FirewallDataSource{}
}

func (d *FirewallDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall"
}

func (d *FirewallDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	fwAttrs := map[string]schema.Attribute{
		"id":                         schema.StringAttribute{Computed: true},
		"name":                       schema.StringAttribute{Computed: true},
		"description":                schema.StringAttribute{Computed: true},
		"customer":                   schema.StringAttribute{Computed: true},
		"customer_name":              schema.StringAttribute{Computed: true},
		"type":                       schema.StringAttribute{Computed: true, Description: "Firewall type: internet or wan."},
		"device":                     schema.StringAttribute{Computed: true},
		"device_name":                schema.StringAttribute{Computed: true},
		"platform":                   schema.StringAttribute{Computed: true},
		"supports_threat_protection": schema.BoolAttribute{Computed: true},
		"context":                    schema.StringAttribute{Computed: true},
		"external_interface":         schema.StringAttribute{Computed: true},
		"internal_interface":         schema.StringAttribute{Computed: true},
		"default_log_profile":        schema.StringAttribute{Computed: true},
		"default_protect_profile":    schema.StringAttribute{Computed: true},
		"multi_tenant":               schema.BoolAttribute{Computed: true},
		"tag_name":                   schema.StringAttribute{Computed: true},
		"nat_ip_sync_enabled":        schema.BoolAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS firewalls. Set `name` or `id` to fetch a single firewall, or omit both to list all (optionally filtered by `customer`).",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact firewall name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific firewall to look up.",
			},
			"customer": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Customer filter for name and list lookups; the customer of the matched firewall.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the matched firewall.",
			},
			"customer_name": schema.StringAttribute{
				Computed:    true,
				Description: "Customer name of the matched firewall.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "Firewall type: internet or wan.",
			},
			"device": schema.StringAttribute{
				Computed:    true,
				Description: "Device of the matched firewall.",
			},
			"device_name": schema.StringAttribute{
				Computed:    true,
				Description: "Device name of the matched firewall.",
			},
			"platform": schema.StringAttribute{
				Computed:    true,
				Description: "Platform of the matched firewall.",
			},
			"supports_threat_protection": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the matched firewall supports threat protection.",
			},
			"context": schema.StringAttribute{
				Computed:    true,
				Description: "VDOM on Fortimanager or Device Group on Panorama.",
			},
			"external_interface": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the external (internet or WAN facing) interface.",
			},
			"internal_interface": schema.StringAttribute{
				Computed:    true,
				Description: "Name of the internal (VDOM or transit facing) interface.",
			},
			"default_log_profile": schema.StringAttribute{
				Computed:    true,
				Description: "Default log profile used when creating firewall rules.",
			},
			"default_protect_profile": schema.StringAttribute{
				Computed:    true,
				Description: "Default protect group profile used when creating firewall rules.",
			},
			"multi_tenant": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the device is multi tenant.",
			},
			"tag_name": schema.StringAttribute{
				Computed:    true,
				Description: "TAG name for object lookups on a multi tenant firewall (PaloAlto only).",
			},
			"nat_ip_sync_enabled": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the daily Panorama NAT public IP import is enabled.",
			},
			"firewalls": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All firewalls (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: fwAttrs},
			},
		},
	}
}

func (d *FirewallDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *FirewallDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config FirewallDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var fw firewallAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/networking/firewalls/%s/", config.Id.ValueString()), &fw)
		if err != nil {
			resp.Diagnostics.AddError("Error reading firewall", err.Error())
			return
		}
		state := FirewallDataSourceModel{FirewallModel: firewallModelFromAPI(&fw), Firewalls: []FirewallModel{}}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	name := config.Name.ValueString()
	query := url.Values{}
	if name != "" {
		query.Set("name__icontains", name)
	}
	if c := config.Customer.ValueString(); c != "" {
		query.Set("customer", c)
	}
	path := "/api/networking/firewalls/"
	if len(query) > 0 {
		path += "?" + query.Encode()
	}

	items, err := listAll[firewallAPIModel](ctx, d.client, path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading firewalls", err.Error())
		return
	}

	if name != "" {
		var match *firewallAPIModel
		for i := range items {
			if items[i].Name == name {
				match = &items[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Firewall not found",
				fmt.Sprintf("No firewall with exact name %q was found.", name))
			return
		}
		state := FirewallDataSourceModel{FirewallModel: firewallModelFromAPI(match), Firewalls: []FirewallModel{}}
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	state := FirewallDataSourceModel{
		FirewallModel: FirewallModel{
			Id:                       types.StringNull(),
			Name:                     types.StringNull(),
			Description:              types.StringNull(),
			Customer:                 config.Customer,
			CustomerName:             types.StringNull(),
			Type:                     types.StringNull(),
			Device:                   types.StringNull(),
			DeviceName:               types.StringNull(),
			Platform:                 types.StringNull(),
			SupportsThreatProtection: types.BoolNull(),
			Context:                  types.StringNull(),
			ExternalInterface:        types.StringNull(),
			InternalInterface:        types.StringNull(),
			DefaultLogProfile:        types.StringNull(),
			DefaultProtectProfile:    types.StringNull(),
			MultiTenant:              types.BoolNull(),
			TagName:                  types.StringNull(),
			NatIpSyncEnabled:         types.BoolNull(),
		},
		Firewalls: make([]FirewallModel, 0, len(items)),
	}
	for i := range items {
		state.Firewalls = append(state.Firewalls, firewallModelFromAPI(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func firewallModelFromAPI(fw *firewallAPIModel) FirewallModel {
	return FirewallModel{
		Id:                       types.StringValue(fw.Id),
		Name:                     types.StringValue(fw.Name),
		Description:              types.StringValue(fw.Description),
		Customer:                 types.StringValue(fw.Customer),
		CustomerName:             types.StringValue(fw.CustomerName),
		Type:                     types.StringValue(fw.Type),
		Device:                   types.StringValue(fw.Device),
		DeviceName:               types.StringValue(fw.DeviceName),
		Platform:                 types.StringValue(fw.Platform),
		SupportsThreatProtection: types.BoolValue(fw.SupportsThreatProtection),
		Context:                  types.StringValue(fw.Context),
		ExternalInterface:        types.StringValue(fw.ExternalInterface),
		InternalInterface:        types.StringValue(fw.InternalInterface),
		DefaultLogProfile:        types.StringValue(fw.DefaultLogProfile),
		DefaultProtectProfile:    types.StringValue(fw.DefaultProtectProfile),
		MultiTenant:              types.BoolValue(fw.MultiTenant),
		TagName:                  types.StringValue(fw.TagName),
		NatIpSyncEnabled:         types.BoolValue(fw.NatIpSyncEnabled),
	}
}
