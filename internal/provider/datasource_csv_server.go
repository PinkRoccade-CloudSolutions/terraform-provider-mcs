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

var _ datasource.DataSource = &CsvServerDataSource{}

type CsvServerDataSource struct {
	client *apiclient.Client
}

type CsvServerDataSourceModel struct {
	Name          types.String     `tfsdk:"name"`
	Id            types.String     `tfsdk:"id"`
	Ufname        types.String     `tfsdk:"ufname"`
	Ipaddress     types.String     `tfsdk:"ipaddress"`
	Port          types.Int64      `tfsdk:"port"`
	Type          types.String     `tfsdk:"type"`
	Policies      types.List       `tfsdk:"policies"`
	Certificate   types.List       `tfsdk:"certificate"`
	CaCertificate types.List       `tfsdk:"ca_certificate"`
	Customer      types.String     `tfsdk:"customer"`
	Loadbalancer  types.String     `tfsdk:"loadbalancer"`
	Clientauth    types.Bool       `tfsdk:"clientauth"`
	Clientcert    types.String     `tfsdk:"clientcert"`
	CsvServers    []CsvServerModel `tfsdk:"csv_servers"`
}

type CsvServerModel struct {
	Id            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	Ufname        types.String `tfsdk:"ufname"`
	Ipaddress     types.String `tfsdk:"ipaddress"`
	Port          types.Int64  `tfsdk:"port"`
	Type          types.String `tfsdk:"type"`
	Policies      types.List   `tfsdk:"policies"`
	Certificate   types.List   `tfsdk:"certificate"`
	CaCertificate types.List   `tfsdk:"ca_certificate"`
	Customer      types.String `tfsdk:"customer"`
	Loadbalancer  types.String `tfsdk:"loadbalancer"`
	Clientauth    types.Bool   `tfsdk:"clientauth"`
	Clientcert    types.String `tfsdk:"clientcert"`
}

type csvServerDSAPIModel struct {
	Id            string   `json:"id"`
	Name          string   `json:"name"`
	Ufname        string   `json:"ufname"`
	Ipaddress     *string  `json:"ipaddress,omitempty"`
	Port          *int64   `json:"port,omitempty"`
	Type          string   `json:"type"`
	Policies      []string `json:"policies,omitempty"`
	Customer      *string  `json:"customer,omitempty"`
	Certificate   []string `json:"certificate,omitempty"`
	CaCertificate []string `json:"ca_certificate,omitempty"`
	Loadbalancer  *string  `json:"loadbalancer,omitempty"`
	Clientauth    *bool    `json:"clientauth,omitempty"`
	Clientcert    *string  `json:"clientcert,omitempty"`
}

func NewCsvServerDataSource() datasource.DataSource {
	return &CsvServerDataSource{}
}

func (d *CsvServerDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_csv_server"
}

func (d *CsvServerDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	csvAttrs := map[string]schema.Attribute{
		"id":             schema.StringAttribute{Computed: true},
		"name":           schema.StringAttribute{Computed: true},
		"ufname":         schema.StringAttribute{Computed: true},
		"ipaddress":      schema.StringAttribute{Computed: true},
		"port":           schema.Int64Attribute{Computed: true},
		"type":           schema.StringAttribute{Computed: true},
		"policies":       schema.ListAttribute{ElementType: types.StringType, Computed: true},
		"certificate":    schema.ListAttribute{ElementType: types.StringType, Computed: true},
		"ca_certificate": schema.ListAttribute{ElementType: types.StringType, Computed: true},
		"customer":       schema.StringAttribute{Computed: true},
		"loadbalancer":   schema.StringAttribute{Computed: true},
		"clientauth":     schema.BoolAttribute{Computed: true},
		"clientcert":     schema.StringAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS CSV servers. Set `name` or `id` to fetch a single CSV server, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact CSV server name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific CSV server to look up.",
			},
			"ufname": schema.StringAttribute{
				Computed:    true,
				Description: "UF name.",
			},
			"ipaddress": schema.StringAttribute{
				Computed:    true,
				Description: "UUID of the associated PublicIPAddress.",
			},
			"port": schema.Int64Attribute{
				Computed:    true,
				Description: "Port number.",
			},
			"type": schema.StringAttribute{
				Computed:    true,
				Description: "CSV server type.",
			},
			"policies": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Bound CS policies.",
			},
			"certificate": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Bound certificates.",
			},
			"ca_certificate": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Bound CA certificates.",
			},
			"customer": schema.StringAttribute{
				Computed:    true,
				Description: "Customer identifier.",
			},
			"loadbalancer": schema.StringAttribute{
				Computed:    true,
				Description: "Associated load balancer.",
			},
			"clientauth": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether client certificate authentication is enabled.",
			},
			"clientcert": schema.StringAttribute{
				Computed:    true,
				Description: "Client certificate requirement (`Mandatory` or `Optional`).",
			},
			"csv_servers": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All CSV servers (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: csvAttrs},
			},
		},
	}
}

func (d *CsvServerDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CsvServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CsvServerDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item csvServerDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/csvserver/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading CSV server", err.Error())
			return
		}
		setSingleCsvServer(ctx, &config, &item, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	items, err := listAll[csvServerDSAPIModel](ctx, d.client, "/api/loadbalancing/csvserver/")
	if err != nil {
		resp.Diagnostics.AddError("Error reading CSV servers", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *csvServerDSAPIModel
		for i := range items {
			if items[i].Name == config.Name.ValueString() {
				match = &items[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("CSV server not found",
				fmt.Sprintf("No CSV server with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleCsvServer(ctx, &config, match, &resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	empty := computedListValue[string](ctx, types.StringType, nil, &resp.Diagnostics)
	state := CsvServerDataSourceModel{
		Name:          types.StringNull(),
		Id:            types.StringNull(),
		Ufname:        types.StringNull(),
		Ipaddress:     types.StringNull(),
		Port:          types.Int64Null(),
		Type:          types.StringNull(),
		Policies:      empty,
		Certificate:   empty,
		CaCertificate: empty,
		Customer:      types.StringNull(),
		Loadbalancer:  types.StringNull(),
		Clientauth:    types.BoolNull(),
		Clientcert:    types.StringNull(),
		CsvServers:    make([]CsvServerModel, 0, len(items)),
	}
	for i := range items {
		state.CsvServers = append(state.CsvServers, csvServerItemToModel(ctx, &items[i], &resp.Diagnostics))
	}
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleCsvServer(ctx context.Context, state *CsvServerDataSourceModel, item *csvServerDSAPIModel, diags *diag.Diagnostics) {
	m := csvServerItemToModel(ctx, item, diags)
	state.Id = m.Id
	state.Name = m.Name
	state.Ufname = m.Ufname
	state.Ipaddress = m.Ipaddress
	state.Port = m.Port
	state.Type = m.Type
	state.Policies = m.Policies
	state.Certificate = m.Certificate
	state.CaCertificate = m.CaCertificate
	state.Customer = m.Customer
	state.Loadbalancer = m.Loadbalancer
	state.Clientauth = m.Clientauth
	state.Clientcert = m.Clientcert
	state.CsvServers = []CsvServerModel{}
}

func csvServerItemToModel(ctx context.Context, item *csvServerDSAPIModel, diags *diag.Diagnostics) CsvServerModel {
	return CsvServerModel{
		Id:            types.StringValue(item.Id),
		Name:          types.StringValue(item.Name),
		Ufname:        types.StringValue(item.Ufname),
		Ipaddress:     types.StringPointerValue(item.Ipaddress),
		Port:          types.Int64PointerValue(item.Port),
		Type:          types.StringValue(item.Type),
		Policies:      computedListValue(ctx, types.StringType, item.Policies, diags),
		Certificate:   computedListValue(ctx, types.StringType, item.Certificate, diags),
		CaCertificate: computedListValue(ctx, types.StringType, item.CaCertificate, diags),
		Customer:      types.StringPointerValue(item.Customer),
		Loadbalancer:  types.StringPointerValue(item.Loadbalancer),
		Clientauth:    types.BoolPointerValue(item.Clientauth),
		Clientcert:    types.StringPointerValue(item.Clientcert),
	}
}
