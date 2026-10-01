package provider

import (
	"context"
	"fmt"
	"net/url"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SecureIngressFirewallDataSource{}

type SecureIngressFirewallDataSource struct {
	client *apiclient.Client
}

type SecureIngressFirewallDataSourceModel struct {
	Id                     types.String                 `tfsdk:"id"`
	Name                   types.String                 `tfsdk:"name"`
	Description            types.String                 `tfsdk:"description"`
	FilterJSON             jsontypes.Normalized         `tfsdk:"filter_json"`
	Customer               types.String                 `tfsdk:"customer"`
	CreatedAtTimestamp     types.String                 `tfsdk:"created_at_timestamp"`
	UpdatedAtTimestamp     types.String                 `tfsdk:"updated_at_timestamp"`
	CreatedByUser          types.Int64                  `tfsdk:"created_by_user"`
	UpdatedByUser          types.Int64                  `tfsdk:"updated_by_user"`
	SecureIngressFirewalls []SecureIngressFirewallModel `tfsdk:"secureingress_firewalls"`
}

type SecureIngressFirewallModel struct {
	Id                 types.String         `tfsdk:"id"`
	Name               types.String         `tfsdk:"name"`
	Description        types.String         `tfsdk:"description"`
	FilterJSON         jsontypes.Normalized `tfsdk:"filter_json"`
	Customer           types.String         `tfsdk:"customer"`
	CreatedAtTimestamp types.String         `tfsdk:"created_at_timestamp"`
	UpdatedAtTimestamp types.String         `tfsdk:"updated_at_timestamp"`
	CreatedByUser      types.Int64          `tfsdk:"created_by_user"`
	UpdatedByUser      types.Int64          `tfsdk:"updated_by_user"`
}

func NewSecureIngressFirewallDataSource() datasource.DataSource {
	return &SecureIngressFirewallDataSource{}
}

func (d *SecureIngressFirewallDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secureingress_firewall"
}

func (d *SecureIngressFirewallDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	itemAttrs := map[string]schema.Attribute{
		"id":                   schema.StringAttribute{Computed: true},
		"name":                 schema.StringAttribute{Computed: true},
		"description":          schema.StringAttribute{Computed: true},
		"filter_json":          schema.StringAttribute{Computed: true, CustomType: jsontypes.NormalizedType{}},
		"customer":             schema.StringAttribute{Computed: true},
		"created_at_timestamp": schema.StringAttribute{Computed: true},
		"updated_at_timestamp": schema.StringAttribute{Computed: true},
		"created_by_user":      schema.Int64Attribute{Computed: true},
		"updated_by_user":      schema.Int64Attribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up secure ingress (XDP) firewalls. Set `name` or `id` to fetch a single firewall, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific firewall to look up.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Exact firewall name to look up.",
			},
			"description": schema.StringAttribute{
				Computed:    true,
				Description: "Description of the firewall.",
			},
			"filter_json": schema.StringAttribute{
				Computed:    true,
				CustomType:  jsontypes.NormalizedType{},
				Description: "Firewall filter definition as a JSON document.",
			},
			"customer": schema.StringAttribute{
				Computed:    true,
				Description: "Customer identifier.",
			},
			"created_at_timestamp": schema.StringAttribute{
				Computed:    true,
				Description: "Time when the firewall was created.",
			},
			"updated_at_timestamp": schema.StringAttribute{
				Computed:    true,
				Description: "Time when the firewall was last updated.",
			},
			"created_by_user": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the user who created the firewall.",
			},
			"updated_by_user": schema.Int64Attribute{
				Computed:    true,
				Description: "ID of the user who last updated the firewall.",
			},
			"secureingress_firewalls": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All secure ingress firewalls (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: itemAttrs},
			},
		},
	}
}

func (d *SecureIngressFirewallDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SecureIngressFirewallDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SecureIngressFirewallDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item secureIngressFirewallAPIModel
		if err := d.client.Get(ctx, fmt.Sprintf("%s%s/", secureIngressFirewallBasePath, config.Id.ValueString()), &item); err != nil {
			resp.Diagnostics.AddError("Error reading secure ingress firewall", err.Error())
			return
		}
		state := secureIngressFirewallSingleState(&item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
		return
	}

	listPath := secureIngressFirewallBasePath
	name := config.Name.ValueString()
	if !config.Name.IsNull() && name != "" {
		listPath += "?name__icontains=" + url.QueryEscape(name)
	}

	items, err := listAll[secureIngressFirewallAPIModel](ctx, d.client, listPath)
	if err != nil {
		resp.Diagnostics.AddError("Error listing secure ingress firewalls", err.Error())
		return
	}

	if !config.Name.IsNull() && name != "" {
		for i := range items {
			if items[i].Name == name {
				state := secureIngressFirewallSingleState(&items[i])
				resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
				return
			}
		}
		resp.Diagnostics.AddError("Secure ingress firewall not found",
			fmt.Sprintf("No secure ingress firewall with exact name %q was found.", name))
		return
	}

	state := SecureIngressFirewallDataSourceModel{
		Id:                     types.StringNull(),
		Name:                   types.StringNull(),
		Description:            types.StringNull(),
		FilterJSON:             jsontypes.NewNormalizedNull(),
		Customer:               types.StringNull(),
		CreatedAtTimestamp:     types.StringNull(),
		UpdatedAtTimestamp:     types.StringNull(),
		CreatedByUser:          types.Int64Null(),
		UpdatedByUser:          types.Int64Null(),
		SecureIngressFirewalls: make([]SecureIngressFirewallModel, 0, len(items)),
	}
	for i := range items {
		state.SecureIngressFirewalls = append(state.SecureIngressFirewalls, secureIngressFirewallItemToModel(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func secureIngressFirewallItemToModel(a *secureIngressFirewallAPIModel) SecureIngressFirewallModel {
	m := SecureIngressFirewallResourceModel{Description: types.StringNull()}
	mapSecureIngressFirewallToState(&m, a)
	return SecureIngressFirewallModel(m)
}

func secureIngressFirewallSingleState(a *secureIngressFirewallAPIModel) SecureIngressFirewallDataSourceModel {
	m := secureIngressFirewallItemToModel(a)
	return SecureIngressFirewallDataSourceModel{
		Id:                     m.Id,
		Name:                   m.Name,
		Description:            m.Description,
		FilterJSON:             m.FilterJSON,
		Customer:               m.Customer,
		CreatedAtTimestamp:     m.CreatedAtTimestamp,
		UpdatedAtTimestamp:     m.UpdatedAtTimestamp,
		CreatedByUser:          m.CreatedByUser,
		UpdatedByUser:          m.UpdatedByUser,
		SecureIngressFirewalls: []SecureIngressFirewallModel{},
	}
}
