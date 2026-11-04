package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &NetworkDhcpLeasesDataSource{}

type NetworkDhcpLeasesDataSource struct {
	client *apiclient.Client
}

type NetworkDhcpLeasesDataSourceModel struct {
	Network types.String         `tfsdk:"network"`
	Ip      types.String         `tfsdk:"ip"`
	Mac     types.String         `tfsdk:"mac"`
	Leases  []dhcpLeaseItemModel `tfsdk:"leases"`
}

type dhcpLeaseItemModel struct {
	Id          types.String `tfsdk:"id"`
	Network     types.String `tfsdk:"network"`
	Ip          types.String `tfsdk:"ip"`
	Mac         types.String `tfsdk:"mac"`
	Description types.String `tfsdk:"description"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func NewNetworkDhcpLeasesDataSource() datasource.DataSource {
	return &NetworkDhcpLeasesDataSource{}
}

func (d *NetworkDhcpLeasesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_network_dhcp_leases"
}

func (d *NetworkDhcpLeasesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists the static DHCP leases of an MCS network, including leases the device sync found on the firewall.",
		Attributes: map[string]schema.Attribute{
			"network": schema.StringAttribute{
				Required:    true,
				Description: "UUID of the network whose leases to list.",
				Validators:  []validator.String{stringvalidator.LengthAtLeast(1)},
			},
			"ip": schema.StringAttribute{
				Optional:    true,
				Description: "Only list the lease for this IPv4 address.",
				Validators:  []validator.String{ipv4Validator{}},
			},
			"mac": schema.StringAttribute{
				Optional:    true,
				Description: "Only list the lease for this MAC address, in any common notation.",
				Validators:  []validator.String{macValidator{}},
			},
			"leases": schema.ListNestedAttribute{
				Computed:    true,
				Description: "The matching leases.",
				NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
					"id":          dsString("UUID of the lease."),
					"network":     dsString("UUID of the network."),
					"ip":          dsString("Reserved IPv4 address."),
					"mac":         dsString("MAC address, lower case and colon separated."),
					"description": dsString("What the reservation is for."),
					"created_at":  dsString("Time the lease was created."),
					"updated_at":  dsString("Time the lease was last updated."),
				}},
			},
		},
	}
}

func (d *NetworkDhcpLeasesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NetworkDhcpLeasesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var cfg NetworkDhcpLeasesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &cfg)...)
	if resp.Diagnostics.HasError() {
		return
	}

	q := url.Values{}
	q.Set("network", cfg.Network.ValueString())
	if known(cfg.Ip) {
		q.Set("ip", cfg.Ip.ValueString())
	}
	if known(cfg.Mac) {
		q.Set("mac", normalizeMAC(cfg.Mac.ValueString()))
	}
	leases, err := listAll[dhcpLeaseAPIModel](ctx, d.client, dhcpLeasesPath+"?"+q.Encode())
	if err != nil {
		resp.Diagnostics.AddError("Error listing network DHCP leases", err.Error())
		return
	}

	cfg.Leases = make([]dhcpLeaseItemModel, 0, len(leases))
	for _, l := range leases {
		cfg.Leases = append(cfg.Leases, dhcpLeaseItemModel{
			Id: types.StringValue(l.Id), Network: types.StringValue(l.Network), Ip: types.StringValue(l.Ip),
			Mac: types.StringValue(l.Mac), Description: types.StringValue(l.Description),
			CreatedAt: types.StringValue(l.CreatedAt), UpdatedAt: types.StringValue(l.UpdatedAt),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &cfg)...)
}
