package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &LbvServerDataSource{}

type LbvServerDataSource struct {
	client *apiclient.Client
}

type LbvServerDataSourceModel struct {
	Name          types.String         `tfsdk:"name"`
	Id            types.String         `tfsdk:"id"`
	Ipaddress     types.String         `tfsdk:"ipaddress"`
	Port          types.Int64          `tfsdk:"port"`
	Type          types.String         `tfsdk:"type"`
	Servicegroup  types.List           `tfsdk:"servicegroup"`
	Certificate   types.List           `tfsdk:"certificate"`
	CaCertificate types.List           `tfsdk:"ca_certificate"`
	Customer      types.String         `tfsdk:"customer"`
	Loadbalancer  types.String         `tfsdk:"loadbalancer"`
	LbvServers    []LbvServerListModel `tfsdk:"lbv_servers"`
}

type LbvServerListModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Ipaddress     types.String `tfsdk:"ipaddress"`
	Port          types.Int64  `tfsdk:"port"`
	Type          types.String `tfsdk:"type"`
	Servicegroup  types.List   `tfsdk:"servicegroup"`
	Certificate   types.List   `tfsdk:"certificate"`
	CaCertificate types.List   `tfsdk:"ca_certificate"`
	Customer      types.String `tfsdk:"customer"`
	Loadbalancer  types.String `tfsdk:"loadbalancer"`
}

type lbvServerDSAPIModel struct {
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	Ipaddress     *string  `json:"ipaddress,omitempty"`
	Port          *int64   `json:"port,omitempty"`
	Type          *string  `json:"type,omitempty"`
	Servicegroup  []string `json:"servicegroup"`
	Certificate   []string `json:"certificate,omitempty"`
	CaCertificate []string `json:"ca_certificate,omitempty"`
	Customer      *string  `json:"customer,omitempty"`
	Loadbalancer  *string  `json:"loadbalancer,omitempty"`
}

func NewLbvServerDataSource() datasource.DataSource {
	return &LbvServerDataSource{}
}

func (d *LbvServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_lbv_server"
}

func (d *LbvServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	srvAttrs := map[string]schema.Attribute{
		"id":             schema.StringAttribute{Computed: true},
		"name":           schema.StringAttribute{Computed: true},
		"ipaddress":      schema.StringAttribute{Computed: true},
		"port":           schema.Int64Attribute{Computed: true},
		"type":           schema.StringAttribute{Computed: true},
		"servicegroup":   schema.ListAttribute{ElementType: types.StringType, Computed: true},
		"certificate":    schema.ListAttribute{ElementType: types.StringType, Computed: true},
		"ca_certificate": schema.ListAttribute{ElementType: types.StringType, Computed: true},
		"customer":       schema.StringAttribute{Computed: true},
		"loadbalancer":   schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS load balancer virtual servers. Set `name` or `id` to fetch a single server, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact virtual server name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific lbv_server to look up.",
			},
			"ipaddress":      schema.StringAttribute{Computed: true},
			"port":           schema.Int64Attribute{Computed: true},
			"type":           schema.StringAttribute{Computed: true},
			"servicegroup":   schema.ListAttribute{ElementType: types.StringType, Computed: true},
			"certificate":    schema.ListAttribute{ElementType: types.StringType, Computed: true},
			"ca_certificate": schema.ListAttribute{ElementType: types.StringType, Computed: true},
			"customer":       schema.StringAttribute{Computed: true},
			"loadbalancer":   schema.StringAttribute{Computed: true},
			"lbv_servers": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All lbv_servers (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: srvAttrs},
			},
		},
	}
}

func (d *LbvServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *LbvServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config LbvServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var s lbvServerDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/lbvserver/%s/", config.Id.ValueString()), &s)
		if err != nil {
			resp.Diagnostics.AddError("Error reading lbv_server", err.Error())
			return
		}
		setSingleLbvServer(ctx, &config, &s, &resp.Diagnostics)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	items, err := listAll[lbvServerDSAPIModel](ctx, d.client, "/api/loadbalancing/lbvserver/")
	if err != nil {
		resp.Diagnostics.AddError("Error reading lbv_servers", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *lbvServerDSAPIModel
		for i := range items {
			if items[i].Name == config.Name.ValueString() {
				match = &items[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("LBV server not found",
				fmt.Sprintf("No lbv_server with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleLbvServer(ctx, &config, match, &resp.Diagnostics)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := LbvServerDataSourceModel{
		Name:          types.StringNull(),
		Id:            types.StringNull(),
		Ipaddress:     types.StringNull(),
		Port:          types.Int64Null(),
		Type:          types.StringNull(),
		Servicegroup:  types.ListNull(types.StringType),
		Certificate:   types.ListNull(types.StringType),
		CaCertificate: types.ListNull(types.StringType),
		Customer:      types.StringNull(),
		Loadbalancer:  types.StringNull(),
		LbvServers:    make([]LbvServerListModel, 0, len(items)),
	}
	for i := range items {
		state.LbvServers = append(state.LbvServers, toLbvServerListModel(ctx, &items[i], &resp.Diagnostics))
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleLbvServer(ctx context.Context, state *LbvServerDataSourceModel, s *lbvServerDSAPIModel, diags *diag.Diagnostics) {
	m := toLbvServerListModel(ctx, s, diags)
	state.Id = m.Id
	state.Name = m.Name
	state.Ipaddress = m.Ipaddress
	state.Port = m.Port
	state.Type = m.Type
	state.Servicegroup = m.Servicegroup
	state.Certificate = m.Certificate
	state.CaCertificate = m.CaCertificate
	state.Customer = m.Customer
	state.Loadbalancer = m.Loadbalancer
	state.LbvServers = []LbvServerListModel{}
}

func toLbvServerListModel(ctx context.Context, s *lbvServerDSAPIModel, diags *diag.Diagnostics) LbvServerListModel {
	return LbvServerListModel{
		Id:            types.StringValue(s.Id),
		Name:          types.StringValue(s.Name),
		Ipaddress:     types.StringPointerValue(s.Ipaddress),
		Port:          types.Int64PointerValue(s.Port),
		Type:          types.StringPointerValue(s.Type),
		Servicegroup:  computedListValue(ctx, types.StringType, s.Servicegroup, diags),
		Certificate:   computedListValue(ctx, types.StringType, s.Certificate, diags),
		CaCertificate: computedListValue(ctx, types.StringType, s.CaCertificate, diags),
		Customer:      types.StringPointerValue(s.Customer),
		Loadbalancer:  types.StringPointerValue(s.Loadbalancer),
	}
}
