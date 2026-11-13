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
	SubjectAltNames  []string           `tfsdk:"subject_alt_names"`
	SubjectAltIps    []string           `tfsdk:"subject_alt_ips"`
	Endpoints        []certEndpointItem `tfsdk:"endpoints"`
	Certificates     []CertificateModel `tfsdk:"certificates"`
}

type CertificateModel struct {
	Id               types.String       `tfsdk:"id"`
	Name             types.String       `tfsdk:"name"`
	Ca               types.Bool         `tfsdk:"ca"`
	ValidToTimestamp types.String       `tfsdk:"valid_to_timestamp"`
	Loadbalancer     types.String       `tfsdk:"loadbalancer"`
	Customer         types.String       `tfsdk:"customer"`
	Protected        types.Bool         `tfsdk:"protected"`
	SubjectAltNames  []string           `tfsdk:"subject_alt_names"`
	SubjectAltIps    []string           `tfsdk:"subject_alt_ips"`
	Endpoints        []certEndpointItem `tfsdk:"endpoints"`
}

// certEndpointItem is the CertificateEndpoint schema: a published endpoint (domain name)
// that uses the certificate or that its SANs cover.
type certEndpointItem struct {
	Kind          types.String `tfsdk:"kind"`
	Hostname      types.String `tfsdk:"hostname"`
	PolicyId      types.String `tfsdk:"policy_id"`
	PolicyName    types.String `tfsdk:"policy_name"`
	CsvserverId   types.String `tfsdk:"csvserver_id"`
	CsvserverName types.String `tfsdk:"csvserver_name"`
	LbvserverId   types.String `tfsdk:"lbvserver_id"`
	LbvserverName types.String `tfsdk:"lbvserver_name"`
	Vip           types.String `tfsdk:"vip"`
	Port          types.Int64  `tfsdk:"port"`
	Bound         types.Bool   `tfsdk:"bound"`
	Covered       types.Bool   `tfsdk:"covered"`
}

type certEndpointAPIModel struct {
	Kind          string  `json:"kind"`
	Hostname      string  `json:"hostname"`
	PolicyId      *string `json:"policy_id"`
	PolicyName    *string `json:"policy_name"`
	CsvserverId   *string `json:"csvserver_id"`
	CsvserverName *string `json:"csvserver_name"`
	LbvserverId   *string `json:"lbvserver_id"`
	LbvserverName *string `json:"lbvserver_name"`
	Vip           *string `json:"vip"`
	Port          *int64  `json:"port"`
	Bound         bool    `json:"bound"`
	Covered       *bool   `json:"covered"`
}

// certificateIncludeEndpoints asks the API to fill `endpoints`, which is null otherwise.
const certificateIncludeEndpoints = "?include=endpoints"

type certificateDSAPIModel struct {
	Id               string                 `json:"id"`
	Name             *string                `json:"name,omitempty"`
	Ca               *bool                  `json:"ca,omitempty"`
	ValidToTimestamp *string                `json:"valid_to_timestamp,omitempty"`
	Loadbalancer     *string                `json:"loadbalancer,omitempty"`
	Customer         *string                `json:"customer,omitempty"`
	Protected        *bool                  `json:"protected,omitempty"`
	SubjectAltNames  []string               `json:"subject_alt_names"`
	SubjectAltIps    []string               `json:"subject_alt_ips"`
	Endpoints        []certEndpointAPIModel `json:"endpoints"`
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
	for k, v := range certificateExtraAttrs() {
		certAttrs[k] = v
	}

	attrs := map[string]schema.Attribute{
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
	}
	for k, v := range certificateExtraAttrs() {
		attrs[k] = v
	}
	resp.Schema = schema.Schema{
		Description: "Look up MCS load balancing certificates. Set `name` or `id` to fetch a single certificate, or omit both to list all.",
		Attributes:  attrs,
	}
}

// certificateExtraAttrs are the SAN and endpoint attributes, shared by the single
// certificate and the `certificates` list.
func certificateExtraAttrs() map[string]schema.Attribute {
	str := func(desc string) schema.Attribute { return schema.StringAttribute{Computed: true, Description: desc} }
	return map[string]schema.Attribute{
		"subject_alt_names": schema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: "DNS names from the certificate's Subject Alternative Name extension.",
		},
		"subject_alt_ips": schema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: "IP addresses from the certificate's Subject Alternative Name extension.",
		},
		"endpoints": schema.ListNestedAttribute{
			Computed:    true,
			Description: "Published endpoints (domain names) that use this certificate or that its SANs cover.",
			NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
				"kind":           str("`http` (hostname route on a content switch) or `l4` (load balancing vserver)."),
				"hostname":       str("Hostname of the endpoint; empty for an L4 endpoint or a vserver without hostname routing."),
				"policy_id":      str("UUID of the content switching policy routing the hostname."),
				"policy_name":    str("Name of the content switching policy."),
				"csvserver_id":   str("UUID of the content switching vserver."),
				"csvserver_name": str("Name of the content switching vserver."),
				"lbvserver_id":   str("UUID of the load balancing vserver."),
				"lbvserver_name": str("Name of the load balancing vserver."),
				"vip":            str("Virtual IP address of the endpoint."),
				"port":           schema.Int64Attribute{Computed: true, Description: "Port of the endpoint."},
				"bound":          schema.BoolAttribute{Computed: true, Description: "Whether the certificate is bound to the vserver serving this endpoint."},
				"covered":        schema.BoolAttribute{Computed: true, Description: "Whether the certificate's SANs cover the hostname; null when unknown (no SANs or no hostname)."},
			}},
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
		err := d.client.Get(ctx, fmt.Sprintf("/api/loadbalancing/certificate/%s/%s", config.Id.ValueString(), certificateIncludeEndpoints), &item)
		if err != nil {
			resp.Diagnostics.AddError("Error reading certificate", err.Error())
			return
		}
		setSingleCertificate(&config, &item)
		resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
		return
	}

	items, err := listAll[certificateDSAPIModel](ctx, d.client, "/api/loadbalancing/certificate/"+certificateIncludeEndpoints)
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
	state.SubjectAltNames = m.SubjectAltNames
	state.SubjectAltIps = m.SubjectAltIps
	state.Endpoints = m.Endpoints
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
		SubjectAltNames:  nonNilStrings(item.SubjectAltNames),
		SubjectAltIps:    nonNilStrings(item.SubjectAltIps),
		Endpoints:        certEndpointsFromAPI(item.Endpoints),
	}
}

func nonNilStrings(xs []string) []string {
	if xs == nil {
		return []string{}
	}
	return xs
}

func certEndpointsFromAPI(eps []certEndpointAPIModel) []certEndpointItem {
	out := make([]certEndpointItem, 0, len(eps))
	for _, e := range eps {
		out = append(out, certEndpointItem{
			Kind:          types.StringValue(e.Kind),
			Hostname:      types.StringValue(e.Hostname),
			PolicyId:      strOrNull(e.PolicyId),
			PolicyName:    strOrNull(e.PolicyName),
			CsvserverId:   strOrNull(e.CsvserverId),
			CsvserverName: strOrNull(e.CsvserverName),
			LbvserverId:   strOrNull(e.LbvserverId),
			LbvserverName: strOrNull(e.LbvserverName),
			Vip:           strOrNull(e.Vip),
			Port:          types.Int64PointerValue(e.Port),
			Bound:         types.BoolValue(e.Bound),
			Covered:       types.BoolPointerValue(e.Covered),
		})
	}
	return out
}
