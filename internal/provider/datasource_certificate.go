package provider

import (
	"context"
	"fmt"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CertificateDataSource{}

type CertificateDataSource struct {
	client *apiclient.Client
}

type CertificateDataSourceModel struct {
	Name             types.String       `tfsdk:"name"`
	Id               types.String       `tfsdk:"id"`
	Ca               types.Bool         `tfsdk:"ca"`
	ValidToTimestamp types.String       `tfsdk:"valid_to_timestamp"`
	Loadbalancer     types.String       `tfsdk:"loadbalancer"`
	Customer         types.String       `tfsdk:"customer"`
	Protected        types.Bool         `tfsdk:"protected"`
	Certificates     []CertificateModel `tfsdk:"certificates"`
}

type CertificateModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Ca               types.Bool   `tfsdk:"ca"`
	ValidToTimestamp types.String `tfsdk:"valid_to_timestamp"`
	Loadbalancer     types.String `tfsdk:"loadbalancer"`
	Customer         types.String `tfsdk:"customer"`
	Protected        types.Bool   `tfsdk:"protected"`
}

type certificateDSAPIModel struct {
	Id               string  `json:"id"`
	Name             *string `json:"name,omitempty"`
	Ca               *bool   `json:"ca,omitempty"`
	ValidToTimestamp *string `json:"valid_to_timestamp,omitempty"`
	Loadbalancer     *string `json:"loadbalancer,omitempty"`
	Customer         *string `json:"customer,omitempty"`
	Protected        *bool   `json:"protected,omitempty"`
}

func NewCertificateDataSource() datasource.DataSource {
	return &CertificateDataSource{}
}

func (d *CertificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (d *CertificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	certAttrs := map[string]schema.Attribute{
		"id":                 schema.StringAttribute{Computed: true},
		"name":               schema.StringAttribute{Computed: true},
		"ca":                 schema.BoolAttribute{Computed: true},
		"valid_to_timestamp": schema.StringAttribute{Computed: true},
		"loadbalancer":       schema.StringAttribute{Computed: true},
		"customer":           schema.StringAttribute{Computed: true},
		"protected":          schema.BoolAttribute{Computed: true},
	}

	resp.Schema = schema.Schema{
		Description: "Look up MCS load balancing certificates. Set `name` or `id` to fetch a single certificate, or omit both to list all.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Exact certificate name to look up.",
			},
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "UUID of a specific certificate to look up.",
			},
			"ca": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the certificate is a CA certificate.",
			},
			"valid_to_timestamp": schema.StringAttribute{
				Computed:    true,
				Description: "Certificate validity end timestamp.",
			},
			"loadbalancer": schema.StringAttribute{
				Computed:    true,
				Description: "Associated load balancer.",
			},
			"customer": schema.StringAttribute{
				Computed:    true,
				Description: "Owning customer.",
			},
			"protected": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether the certificate is protected from changes.",
			},
			"certificates": schema.ListNestedAttribute{
				Computed:     true,
				Description:  "All certificates (populated when neither `name` nor `id` is set).",
				NestedObject: schema.NestedAttributeObject{Attributes: certAttrs},
			},
		},
	}
}

func (d *CertificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CertificateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !config.Id.IsNull() && config.Id.ValueString() != "" {
		var item certificateDSAPIModel
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/certificate/%s/", config.Id.ValueString()), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading certificate", err.Error())
			return
		}
		setSingleCertificate(&config, &item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	items, err := listAll[certificateDSAPIModel](ctx, d.client, "/api/loadbalancing/certificate/")
	if err != nil {
		resp.Diagnostics.AddError("Error reading certificates", err.Error())
		return
	}

	if !config.Name.IsNull() && config.Name.ValueString() != "" {
		var match *certificateDSAPIModel
		for i := range items {
			if items[i].Name != nil && *items[i].Name == config.Name.ValueString() {
				match = &items[i]
				break
			}
		}
		if match == nil {
			resp.Diagnostics.AddError("Certificate not found",
				fmt.Sprintf("No certificate with exact name %q was found.", config.Name.ValueString()))
			return
		}
		setSingleCertificate(&config, match)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	state := CertificateDataSourceModel{
		Name:             types.StringNull(),
		Id:               types.StringNull(),
		Ca:               types.BoolNull(),
		ValidToTimestamp: types.StringNull(),
		Loadbalancer:     types.StringNull(),
		Customer:         types.StringNull(),
		Protected:        types.BoolNull(),
		Certificates:     make([]CertificateModel, 0, len(items)),
	}
	for i := range items {
		state.Certificates = append(state.Certificates, certificateItemToModel(&items[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func setSingleCertificate(state *CertificateDataSourceModel, item *certificateDSAPIModel) {
	m := certificateItemToModel(item)
	state.Id = m.Id
	state.Name = m.Name
	state.Ca = m.Ca
	state.ValidToTimestamp = m.ValidToTimestamp
	state.Loadbalancer = m.Loadbalancer
	state.Customer = m.Customer
	state.Protected = m.Protected
	state.Certificates = []CertificateModel{}
}

func certificateItemToModel(item *certificateDSAPIModel) CertificateModel {
	return CertificateModel{
		Id:               types.StringValue(item.Id),
		Name:             types.StringPointerValue(item.Name),
		Ca:               types.BoolPointerValue(item.Ca),
		ValidToTimestamp: types.StringPointerValue(item.ValidToTimestamp),
		Loadbalancer:     types.StringPointerValue(item.Loadbalancer),
		Customer:         types.StringPointerValue(item.Customer),
		Protected:        types.BoolPointerValue(item.Protected),
	}
}
