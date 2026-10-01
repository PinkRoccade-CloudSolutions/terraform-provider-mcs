package provider

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SiteToSiteVPNDataSource{}

type SiteToSiteVPNDataSource struct {
	client *apiclient.Client
}

type SiteToSiteVPNDataSourceModel struct {
	Id                 types.String             `tfsdk:"id"`
	Name               types.String             `tfsdk:"name"`
	Uuid               types.String             `tfsdk:"uuid"`
	Domain             types.Int64              `tfsdk:"domain"`
	Tenant             types.Int64              `tfsdk:"tenant"`
	Customer           types.String             `tfsdk:"customer"`
	Firewall           types.String             `tfsdk:"firewall"`
	Contact            types.List               `tfsdk:"contact"`
	State              types.String             `tfsdk:"state"`
	LastStatus         types.String             `tfsdk:"last_status"`
	Resets             types.Int64              `tfsdk:"resets"`
	LastCheck          types.String             `tfsdk:"last_check"`
	LastReset          types.String             `tfsdk:"last_reset"`
	CreatedAtTimestamp types.String             `tfsdk:"created_at_timestamp"`
	UpdatedAtTimestamp types.String             `tfsdk:"updated_at_timestamp"`
	CreatedByUser      types.Int64              `tfsdk:"created_by_user"`
	UpdatedByUser      types.Int64              `tfsdk:"updated_by_user"`
	Vpns               []SiteToSiteVPNListModel `tfsdk:"vpns"`
}

type SiteToSiteVPNListModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Uuid               types.String `tfsdk:"uuid"`
	Domain             types.Int64  `tfsdk:"domain"`
	Tenant             types.Int64  `tfsdk:"tenant"`
	Customer           types.String `tfsdk:"customer"`
	Firewall           types.String `tfsdk:"firewall"`
	Contact            types.List   `tfsdk:"contact"`
	State              types.String `tfsdk:"state"`
	LastStatus         types.String `tfsdk:"last_status"`
	Resets             types.Int64  `tfsdk:"resets"`
	LastCheck          types.String `tfsdk:"last_check"`
	LastReset          types.String `tfsdk:"last_reset"`
	CreatedAtTimestamp types.String `tfsdk:"created_at_timestamp"`
	UpdatedAtTimestamp types.String `tfsdk:"updated_at_timestamp"`
	CreatedByUser      types.Int64  `tfsdk:"created_by_user"`
	UpdatedByUser      types.Int64  `tfsdk:"updated_by_user"`
}

type siteToSiteVPNDSAPIModel struct {
	Id                 int     `json:"id"`
	Uuid               string  `json:"uuid"`
	Name               string  `json:"name"`
	Domain             *int64  `json:"domain"`
	Tenant             *int64  `json:"tenant"`
	Customer           *string `json:"customer"`
	Firewall           *string `json:"firewall"`
	Contact            []int64 `json:"contact"`
	State              string  `json:"state"`
	LastStatus         string  `json:"last_status"`
	Resets             int64   `json:"resets"`
	LastCheck          *string `json:"last_check"`
	LastReset          *string `json:"last_reset"`
	CreatedAtTimestamp string  `json:"created_at_timestamp"`
	UpdatedAtTimestamp string  `json:"updated_at_timestamp"`
	CreatedByUser      *int64  `json:"created_by_user"`
	UpdatedByUser      *int64  `json:"updated_by_user"`
}

func NewSiteToSiteVPNDataSource() datasource.DataSource {
	return &SiteToSiteVPNDataSource{}
}

func (d *SiteToSiteVPNDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_site_to_site_vpn"
}

func siteToSiteVPNComputedAttrs() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"uuid":                 schema.StringAttribute{Computed: true},
		"domain":               schema.Int64Attribute{Computed: true},
		"tenant":               schema.Int64Attribute{Computed: true},
		"customer":             schema.StringAttribute{Computed: true},
		"firewall":             schema.StringAttribute{Computed: true},
		"contact":              schema.ListAttribute{Computed: true, ElementType: types.Int64Type},
		"state":                schema.StringAttribute{Computed: true},
		"last_status":          schema.StringAttribute{Computed: true},
		"resets":               schema.Int64Attribute{Computed: true},
		"last_check":           schema.StringAttribute{Computed: true},
		"last_reset":           schema.StringAttribute{Computed: true},
		"created_at_timestamp": schema.StringAttribute{Computed: true},
		"updated_at_timestamp": schema.StringAttribute{Computed: true},
		"created_by_user":      schema.Int64Attribute{Computed: true},
		"updated_by_user":      schema.Int64Attribute{Computed: true},
	}
}

func (d *SiteToSiteVPNDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	vpnAttrs := siteToSiteVPNComputedAttrs()
	vpnAttrs["id"] = schema.StringAttribute{Computed: true}
	vpnAttrs["name"] = schema.StringAttribute{Computed: true}

	attrs := siteToSiteVPNComputedAttrs()
	attrs["name"] = schema.StringAttribute{
		Optional:    true,
		Description: "Exact VPN name to look up.",
	}
	attrs["id"] = schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Description: "Numeric API id of a specific VPN to look up, or the id of the matched VPN when filtering by name.",
	}
	attrs["vpns"] = schema.ListNestedAttribute{
		Computed:     true,
		Description:  "All site-to-site VPNs (populated when neither `name` nor `id` is set).",
		NestedObject: schema.NestedAttributeObject{Attributes: vpnAttrs},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS site-to-site VPNs. Set `name` or `id` to fetch a single VPN, or omit both to list all.",
		Attributes:  attrs,
	}
}

func (d *SiteToSiteVPNDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SiteToSiteVPNDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SiteToSiteVPNDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var vpn siteToSiteVPNDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/vpn/site_to_site/%s/", config.Id.ValueString()), &vpn)
		if err != nil {
			resp.Diagnostics.AddError("Error reading site-to-site VPN", err.Error())
			return
		}
		state := singleSiteToSiteVPN(toSiteToSiteVPNListModel(ctx, &vpn, &resp.Diagnostics))
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	path := "/api/vpn/site_to_site/"
	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		path += "?name=" + url.QueryEscape(config.Name.ValueString())
	}

	items, err := listAll[siteToSiteVPNDSAPIModel](ctx, d.client, path)
	if err != nil {
		resp.Diagnostics.AddError("Error reading site-to-site VPNs", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *siteToSiteVPNDSAPIModel
		for i := range items {
			if items[i].Name == config.Name.ValueString() {
				match = &items[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Site-to-site VPN not found",
				fmt.Sprintf("No site-to-site VPN with exact name %q was found.", config.Name.ValueString()))
			return
		}
		state := singleSiteToSiteVPN(toSiteToSiteVPNListModel(ctx, match, &resp.Diagnostics))
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	state := SiteToSiteVPNDataSourceModel{
		Id:                 types.StringNull(),
		Name:               types.StringNull(),
		Uuid:               types.StringNull(),
		Domain:             types.Int64Null(),
		Tenant:             types.Int64Null(),
		Customer:           types.StringNull(),
		Firewall:           types.StringNull(),
		Contact:            types.ListNull(types.Int64Type),
		State:              types.StringNull(),
		LastStatus:         types.StringNull(),
		Resets:             types.Int64Null(),
		LastCheck:          types.StringNull(),
		LastReset:          types.StringNull(),
		CreatedAtTimestamp: types.StringNull(),
		UpdatedAtTimestamp: types.StringNull(),
		CreatedByUser:      types.Int64Null(),
		UpdatedByUser:      types.Int64Null(),
		Vpns:               make([]SiteToSiteVPNListModel, 0, len(items)),
	}
	for i := range items {
		state.Vpns = append(state.Vpns, toSiteToSiteVPNListModel(ctx, &items[i], &resp.Diagnostics))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func singleSiteToSiteVPN(m SiteToSiteVPNListModel) SiteToSiteVPNDataSourceModel {
	return SiteToSiteVPNDataSourceModel{
		Id:                 m.Id,
		Name:               m.Name,
		Uuid:               m.Uuid,
		Domain:             m.Domain,
		Tenant:             m.Tenant,
		Customer:           m.Customer,
		Firewall:           m.Firewall,
		Contact:            m.Contact,
		State:              m.State,
		LastStatus:         m.LastStatus,
		Resets:             m.Resets,
		LastCheck:          m.LastCheck,
		LastReset:          m.LastReset,
		CreatedAtTimestamp: m.CreatedAtTimestamp,
		UpdatedAtTimestamp: m.UpdatedAtTimestamp,
		CreatedByUser:      m.CreatedByUser,
		UpdatedByUser:      m.UpdatedByUser,
		Vpns:               []SiteToSiteVPNListModel{},
	}
}

func toSiteToSiteVPNListModel(ctx context.Context, vpn *siteToSiteVPNDSAPIModel, diags *diag.Diagnostics) SiteToSiteVPNListModel {
	return SiteToSiteVPNListModel{
		Id:                 types.StringValue(strconv.Itoa(vpn.Id)),
		Name:               types.StringValue(vpn.Name),
		Uuid:               types.StringValue(vpn.Uuid),
		Domain:             types.Int64PointerValue(vpn.Domain),
		Tenant:             types.Int64PointerValue(vpn.Tenant),
		Customer:           types.StringPointerValue(vpn.Customer),
		Firewall:           types.StringPointerValue(vpn.Firewall),
		Contact:            computedListValue(ctx, types.Int64Type, vpn.Contact, diags),
		State:              types.StringValue(vpn.State),
		LastStatus:         types.StringValue(vpn.LastStatus),
		Resets:             types.Int64Value(vpn.Resets),
		LastCheck:          types.StringPointerValue(vpn.LastCheck),
		LastReset:          types.StringPointerValue(vpn.LastReset),
		CreatedAtTimestamp: types.StringValue(vpn.CreatedAtTimestamp),
		UpdatedAtTimestamp: types.StringValue(vpn.UpdatedAtTimestamp),
		CreatedByUser:      types.Int64PointerValue(vpn.CreatedByUser),
		UpdatedByUser:      types.Int64PointerValue(vpn.UpdatedByUser),
	}
}
