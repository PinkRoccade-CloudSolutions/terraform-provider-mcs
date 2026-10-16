package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &SecureIngressRouteDataSource{}

func NewSecureIngressRouteDataSource() datasource.DataSource {
	return &SecureIngressRouteDataSource{}
}

type SecureIngressRouteDataSource struct {
	client *apiclient.Client
}

type SecureIngressRouteDataSourceModel struct {
	Id         types.String     `tfsdk:"id"`
	Cluster    types.String     `tfsdk:"cluster"`
	ServerName types.String     `tfsdk:"server_name"`
	Routes     []routeItemModel `tfsdk:"routes"`
}

type routeItemModel struct {
	Id                          types.String `tfsdk:"id"`
	Cluster                     types.String `tfsdk:"cluster"`
	ServerName                  types.String `tfsdk:"server_name"`
	TargetKind                  types.String `tfsdk:"target_kind"`
	Virtualmachine              types.String `tfsdk:"virtualmachine"`
	Interface                   types.String `tfsdk:"interface"`
	KubernetesCluster           types.String `tfsdk:"kubernetes_cluster"`
	UpstreamAddress             types.String `tfsdk:"upstream_address"`
	Port                        types.Int64  `tfsdk:"port"`
	Scheme                      types.String `tfsdk:"scheme"`
	Path                        types.String `tfsdk:"path"`
	Ssl                         types.Bool   `tfsdk:"ssl"`
	Upstream                    types.String `tfsdk:"upstream"`
	CertificateCommonName       types.String `tfsdk:"certificate_common_name"`
	CertificateIssuer           types.String `tfsdk:"certificate_issuer"`
	CertificateValidToTimestamp types.String `tfsdk:"certificate_valid_to_timestamp"`
}

func (d *SecureIngressRouteDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secureingress_route"
}

func (d *SecureIngressRouteDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dschema.Schema{
		Description: "Lists secure ingress routes, optionally filtered by ID, ingress cluster and server name.",
		Attributes: map[string]dschema.Attribute{
			"id":          dschema.StringAttribute{Optional: true, Description: "Return only the route with this ID."},
			"cluster":     dschema.StringAttribute{Optional: true, Description: "Return only routes of this ingress cluster (UUID)."},
			"server_name": dschema.StringAttribute{Optional: true, Description: "Return only the route with this exact server name."},
			"routes": dschema.ListNestedAttribute{
				Computed:    true,
				Description: "Matching routes.",
				NestedObject: dschema.NestedAttributeObject{Attributes: map[string]dschema.Attribute{
					"id":                             dsString("Route ID."),
					"cluster":                        dsString("Ingress cluster UUID."),
					"server_name":                    dsString("Hostname served by the route."),
					"target_kind":                    dsString("Upstream kind: `vm`, `kubernetes` or `address`."),
					"virtualmachine":                 dsString("Upstream virtual machine UUID."),
					"interface":                      dsString("Upstream interface UUID."),
					"kubernetes_cluster":             dsString("Upstream Kubernetes cluster UUID."),
					"upstream_address":               dsString("Upstream IP address."),
					"port":                           dsInt64("Upstream port."),
					"scheme":                         dsString("Upstream scheme."),
					"path":                           dsString("Path the route serves."),
					"ssl":                            dsBool("Whether the route is served over HTTPS."),
					"upstream":                       dsString("Resolved upstream URL."),
					"certificate_common_name":        dsString("Common name of the custom certificate, if any."),
					"certificate_issuer":             dsString("Issuer of the custom certificate, if any."),
					"certificate_valid_to_timestamp": dsString("End of validity of the custom certificate, if any."),
				}},
			},
		},
	}
}

func (d *SecureIngressRouteDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func routeToItem(a *routeAPIModel) routeItemModel {
	item := routeItemModel{
		Id:                          types.StringValue(a.ID),
		Cluster:                     types.StringValue(a.Cluster),
		ServerName:                  types.StringValue(a.ServerName),
		TargetKind:                  types.StringValue(a.TargetKind),
		Virtualmachine:              strOrNull(a.Virtualmachine),
		Interface:                   strOrNull(a.Interface),
		KubernetesCluster:           strOrNull(a.KubernetesCluster),
		UpstreamAddress:             types.StringValue(a.UpstreamAddress),
		Port:                        types.Int64Value(a.Port),
		Scheme:                      types.StringValue(a.Scheme),
		Path:                        types.StringValue(a.Path),
		Ssl:                         types.BoolValue(a.Ssl),
		Upstream:                    types.StringValue(a.Upstream),
		CertificateCommonName:       types.StringNull(),
		CertificateIssuer:           types.StringNull(),
		CertificateValidToTimestamp: types.StringNull(),
	}
	if c := a.CustomCertificate; c != nil {
		item.CertificateCommonName = types.StringValue(c.CommonName)
		item.CertificateIssuer = types.StringValue(c.Issuer)
		item.CertificateValidToTimestamp = strOrNull(c.ValidToTimestamp)
	}
	return item
}

func (d *SecureIngressRouteDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SecureIngressRouteDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var routes []routeAPIModel
	if known(config.Id) {
		var route routeAPIModel
		if err := d.client.Get(ctx, routeCollection+url.PathEscape(config.Id.ValueString())+"/", &route); err != nil {
			resp.Diagnostics.AddError("Error reading secure ingress route", err.Error())
			return
		}
		routes = append(routes, route)
	} else {
		params := url.Values{}
		if known(config.Cluster) {
			params.Set("cluster", config.Cluster.ValueString())
		}
		if known(config.ServerName) {
			params.Set("server_name__icontains", config.ServerName.ValueString())
		}
		listPath := routeCollection
		if len(params) > 0 {
			listPath += "?" + params.Encode()
		}
		raw, err := d.client.ListAll(ctx, listPath)
		if err != nil {
			resp.Diagnostics.AddError("Error listing secure ingress routes", err.Error())
			return
		}
		for _, r := range raw {
			var route routeAPIModel
			if err := json.Unmarshal(r, &route); err != nil {
				resp.Diagnostics.AddError("Error parsing secure ingress route", err.Error())
				return
			}
			routes = append(routes, route)
		}
	}

	config.Routes = []routeItemModel{}
	for i := range routes {
		a := &routes[i]
		if known(config.Cluster) && a.Cluster != config.Cluster.ValueString() {
			continue
		}
		if known(config.ServerName) && !strings.EqualFold(a.ServerName, config.ServerName.ValueString()) {
			continue
		}
		config.Routes = append(config.Routes, routeToItem(a))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}
