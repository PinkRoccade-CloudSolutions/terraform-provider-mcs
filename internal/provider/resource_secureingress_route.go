package provider

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/PinkRoccade-CloudSolutions/terraform-provider-mcs/internal/apiclient"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const routeCollection = "/api/secureingress/route/"

var (
	_ resource.Resource                = &SecureIngressRouteResource{}
	_ resource.ResourceWithConfigure   = &SecureIngressRouteResource{}
	_ resource.ResourceWithImportState = &SecureIngressRouteResource{}
)

func NewSecureIngressRouteResource() resource.Resource {
	return &SecureIngressRouteResource{}
}

type SecureIngressRouteResource struct {
	client *apiclient.Client
}

type SecureIngressRouteResourceModel struct {
	Id                types.String         `tfsdk:"id"`
	Cluster           types.String         `tfsdk:"cluster"`
	ServerName        types.String         `tfsdk:"server_name"`
	TargetKind        types.String         `tfsdk:"target_kind"`
	Virtualmachine    types.String         `tfsdk:"virtualmachine"`
	Interface         types.String         `tfsdk:"interface"`
	KubernetesCluster types.String         `tfsdk:"kubernetes_cluster"`
	UpstreamAddress   types.String         `tfsdk:"upstream_address"`
	Port              types.Int64          `tfsdk:"port"`
	Scheme            types.String         `tfsdk:"scheme"`
	Path              types.String         `tfsdk:"path"`
	Ssl               types.Bool           `tfsdk:"ssl"`
	Settings          types.Map            `tfsdk:"settings"`
	SecretSettings    types.Map            `tfsdk:"secret_settings"`
	SecretsSet        types.List           `tfsdk:"secrets_set"`
	Upstream          types.String         `tfsdk:"upstream"`
	Locations         []routeLocationModel `tfsdk:"locations"`
	WafExceptions     []routeWafModel      `tfsdk:"waf_exceptions"`
	CustomCertificate *routeCertModel      `tfsdk:"custom_certificate"`
}

type routeLocationModel struct {
	Id                types.String `tfsdk:"id"`
	Slot              types.Int64  `tfsdk:"slot"`
	Path              types.String `tfsdk:"path"`
	TargetKind        types.String `tfsdk:"target_kind"`
	Virtualmachine    types.String `tfsdk:"virtualmachine"`
	Interface         types.String `tfsdk:"interface"`
	KubernetesCluster types.String `tfsdk:"kubernetes_cluster"`
	UpstreamAddress   types.String `tfsdk:"upstream_address"`
	Scheme            types.String `tfsdk:"scheme"`
	Port              types.Int64  `tfsdk:"port"`
	Settings          types.Map    `tfsdk:"settings"`
	Upstream          types.String `tfsdk:"upstream"`
}

type routeWafModel struct {
	Id        types.Int64  `tfsdk:"id"`
	Kind      types.String `tfsdk:"kind"`
	Rules     types.String `tfsdk:"rules"`
	Path      types.String `tfsdk:"path"`
	PathMatch types.String `tfsdk:"path_match"`
	Target    types.String `tfsdk:"target"`
	Raw       types.String `tfsdk:"raw"`
	Reason    types.String `tfsdk:"reason"`
}

type routeCertModel struct {
	Format             types.String `tfsdk:"format"`
	Certificate        types.String `tfsdk:"certificate"`
	PrivateKey         types.String `tfsdk:"private_key"`
	Passphrase         types.String `tfsdk:"passphrase"`
	Id                 types.String `tfsdk:"id"`
	CommonName         types.String `tfsdk:"common_name"`
	SubjectAltNames    types.List   `tfsdk:"subject_alt_names"`
	Issuer             types.String `tfsdk:"issuer"`
	SerialNumber       types.String `tfsdk:"serial_number"`
	FingerprintSha256  types.String `tfsdk:"fingerprint_sha256"`
	ValidFromTimestamp types.String `tfsdk:"valid_from_timestamp"`
	ValidToTimestamp   types.String `tfsdk:"valid_to_timestamp"`
}

// --- API representation ----------------------------------------------------

// settingsMap decodes setting values as strings; non-string JSON values keep
// their JSON text.
type settingsMap map[string]string

func (s *settingsMap) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	out := make(settingsMap, len(raw))
	for k, v := range raw {
		var str string
		if err := json.Unmarshal(v, &str); err == nil {
			out[k] = str
		} else if string(v) != "null" {
			out[k] = string(v)
		}
	}
	*s = out
	return nil
}

type routeAPIModel struct {
	ID                string             `json:"id"`
	Cluster           string             `json:"cluster"`
	ServerName        string             `json:"server_name"`
	TargetKind        string             `json:"target_kind"`
	Virtualmachine    *string            `json:"virtualmachine"`
	Interface         *string            `json:"interface"`
	KubernetesCluster *string            `json:"kubernetes_cluster"`
	UpstreamAddress   string             `json:"upstream_address"`
	Port              int64              `json:"port"`
	Scheme            string             `json:"scheme"`
	Path              string             `json:"path"`
	Ssl               bool               `json:"ssl"`
	CustomCertificate *routeCertAPIModel `json:"custom_certificate"`
	Settings          settingsMap        `json:"settings"`
	SecretsSet        []string           `json:"secrets_set"`
	Locations         []routeLocationAPI `json:"locations"`
	Upstream          string             `json:"upstream"`
	WafExceptions     []routeWafAPIModel `json:"waf_exceptions"`
}

type routeLocationAPI struct {
	ID                string      `json:"id"`
	Slot              int64       `json:"slot"`
	Path              string      `json:"path"`
	TargetKind        string      `json:"target_kind"`
	Virtualmachine    *string     `json:"virtualmachine"`
	Interface         *string     `json:"interface"`
	KubernetesCluster *string     `json:"kubernetes_cluster"`
	UpstreamAddress   string      `json:"upstream_address"`
	Scheme            string      `json:"scheme"`
	Port              int64       `json:"port"`
	Settings          settingsMap `json:"settings"`
	Upstream          string      `json:"upstream"`
}

type routeWafAPIModel struct {
	ID        int64  `json:"id"`
	Kind      string `json:"kind"`
	Rules     string `json:"rules"`
	Path      string `json:"path"`
	PathMatch string `json:"path_match"`
	Target    string `json:"target"`
	Raw       string `json:"raw"`
	Reason    string `json:"reason"`
}

type routeCertAPIModel struct {
	ID                 string   `json:"id"`
	CommonName         string   `json:"common_name"`
	SubjectAltNames    []string `json:"subject_alt_names"`
	Issuer             string   `json:"issuer"`
	SerialNumber       string   `json:"serial_number"`
	FingerprintSha256  string   `json:"fingerprint_sha256"`
	ValidFromTimestamp *string  `json:"valid_from_timestamp"`
	ValidToTimestamp   *string  `json:"valid_to_timestamp"`
}

type routeBody struct {
	Cluster           *string              `json:"cluster,omitempty"`
	ServerName        string               `json:"server_name"`
	TargetKind        *string              `json:"target_kind,omitempty"`
	Virtualmachine    *string              `json:"virtualmachine"`
	Interface         *string              `json:"interface"`
	KubernetesCluster *string              `json:"kubernetes_cluster"`
	UpstreamAddress   *string              `json:"upstream_address,omitempty"`
	Port              int64                `json:"port"`
	Scheme            string               `json:"scheme"`
	Path              string               `json:"path"`
	Ssl               bool                 `json:"ssl"`
	Settings          map[string]*string   `json:"settings,omitempty"`
	Locations         *[]routeLocationBody `json:"locations,omitempty"`
	WafExceptions     *[]routeWafBody      `json:"waf_exceptions,omitempty"`
}

type routeLocationBody struct {
	ID                *string           `json:"id,omitempty"`
	Path              string            `json:"path"`
	TargetKind        *string           `json:"target_kind,omitempty"`
	Virtualmachine    *string           `json:"virtualmachine"`
	Interface         *string           `json:"interface"`
	KubernetesCluster *string           `json:"kubernetes_cluster"`
	UpstreamAddress   *string           `json:"upstream_address,omitempty"`
	Scheme            string            `json:"scheme"`
	Port              int64             `json:"port"`
	Settings          map[string]string `json:"settings"`
}

type routeWafBody struct {
	ID        *int64  `json:"id,omitempty"`
	Kind      string  `json:"kind"`
	Rules     *string `json:"rules,omitempty"`
	Path      *string `json:"path,omitempty"`
	PathMatch *string `json:"path_match,omitempty"`
	Target    *string `json:"target,omitempty"`
	Raw       *string `json:"raw,omitempty"`
	Reason    string  `json:"reason"`
}

// --- Schema ------------------------------------------------------------------

func (r *SecureIngressRouteResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_secureingress_route"
}

func targetAttrs(level string) map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"target_kind": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Upstream kind: `vm`, `kubernetes` or `address`. Inferred from the target fields when omitted.",
		},
		"virtualmachine":     rsOptionalString("UUID of the upstream virtual machine (`vm` target)."),
		"interface":          rsOptionalString("UUID of the virtual machine interface to proxy to (`vm` target)."),
		"kubernetes_cluster": rsOptionalString("UUID of the upstream Kubernetes cluster (`kubernetes` target)."),
		"upstream_address": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Upstream IP address. Required for an `address` target, optional for `kubernetes`, and derived from the interface for `vm`.",
		},
		"port": schema.Int64Attribute{Required: true, Description: "Upstream port (1-65535)."},
		"scheme": schema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Default:     stringdefault.StaticString("http"),
			Description: "Upstream scheme: `http` or `https`. Defaults to `http`.",
		},
		"settings": schema.MapAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "BunkerWeb setting overrides for this " + level + ". Values are strings; booleans are `\"yes\"`/`\"no\"`.",
		},
		"upstream": rsVolatileString("Resolved upstream URL."),
	}
}

func (r *SecureIngressRouteResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	attrs := targetAttrs("route")
	attrs["id"] = rsID()
	attrs["cluster"] = schema.StringAttribute{
		Required:      true,
		Description:   "UUID of the ingress cluster. Changing it forces a new route.",
		PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
	}
	attrs["server_name"] = rsRequiredString("Fully qualified hostname served by the route, e.g. `shop.example.com`.")
	attrs["path"] = schema.StringAttribute{
		Optional:    true,
		Computed:    true,
		Default:     stringdefault.StaticString("/"),
		Description: "Path the route serves. Defaults to `/`.",
	}
	attrs["ssl"] = schema.BoolAttribute{
		Optional:    true,
		Computed:    true,
		Default:     booldefault.StaticBool(true),
		Description: "Serve the route over HTTPS. Defaults to `true`.",
	}
	attrs["secret_settings"] = schema.MapAttribute{
		ElementType: types.StringType,
		Optional:    true,
		Sensitive:   true,
		Description: "Write-only BunkerWeb settings such as `AUTH_BASIC_PASSWORD`. The API never returns them; `secrets_set` lists which are stored.",
	}
	attrs["secrets_set"] = schema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
		Description: "Write-only settings that have a stored value.",
	}

	locAttrs := targetAttrs("location")
	locAttrs["id"] = schema.StringAttribute{Computed: true, Description: "Location ID."}
	locAttrs["slot"] = schema.Int64Attribute{Computed: true, Description: "BunkerWeb slot number of the location."}
	locAttrs["path"] = rsRequiredString("Path of the location. Must start with `/` and not be `/`.")
	attrs["locations"] = schema.ListNestedAttribute{
		Optional:     true,
		Description:  "Extra paths of the route with their own upstream. Locations are matched to existing ones by `path`.",
		NestedObject: schema.NestedAttributeObject{Attributes: locAttrs},
	}

	attrs["waf_exceptions"] = schema.ListNestedAttribute{
		Optional:    true,
		Description: "WAF exceptions, applied in list order. Existing exceptions are matched by position.",
		NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{Computed: true, Description: "Exception ID."},
			"kind": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("structured"),
				Description: "`structured` or `raw` (raw ModSecurity rules need extra permissions). Defaults to `structured`.",
			},
			"rules":  schema.StringAttribute{Optional: true, Computed: true, Description: "ModSecurity rule IDs to disable (structured)."},
			"path":   schema.StringAttribute{Optional: true, Computed: true, Description: "Path the exception applies to (structured)."},
			"target": schema.StringAttribute{Optional: true, Computed: true, Description: "Request part to exclude, e.g. `ARGS:q` (structured)."},
			"path_match": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "`exact` or `prefix` (structured). Defaults to `exact`.",
			},
			"raw":    schema.StringAttribute{Optional: true, Computed: true, Description: "Raw ModSecurity configuration (raw)."},
			"reason": rsRequiredString("Why the exception is needed."),
		}},
	}

	attrs["custom_certificate"] = schema.SingleNestedAttribute{
		Optional:    true,
		Description: "Custom certificate for the route. Without it the route uses Let's Encrypt; removing it reverts to Let's Encrypt.",
		Attributes: map[string]schema.Attribute{
			"format": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("PEM"),
				Description: "`PEM`, `PFX` or `P7B`. Defaults to `PEM`.",
			},
			"certificate": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Certificate content. PEM text (with intermediates), or base64 (e.g. `filebase64()`) for a binary PFX/P7B bundle.",
			},
			"private_key": schema.StringAttribute{Optional: true, Sensitive: true, Description: "PEM private key (PEM and P7B formats)."},
			"passphrase":  schema.StringAttribute{Optional: true, Sensitive: true, Description: "Passphrase of the PFX bundle or encrypted key."},
			"id":          schema.StringAttribute{Computed: true, Description: "Certificate ID."},
			"common_name": schema.StringAttribute{Computed: true, Description: "Certificate common name."},
			"subject_alt_names": schema.ListAttribute{
				ElementType: types.StringType, Computed: true, Description: "Subject alternative names.",
			},
			"issuer":               schema.StringAttribute{Computed: true, Description: "Certificate issuer."},
			"serial_number":        schema.StringAttribute{Computed: true, Description: "Certificate serial number."},
			"fingerprint_sha256":   schema.StringAttribute{Computed: true, Description: "SHA-256 fingerprint."},
			"valid_from_timestamp": schema.StringAttribute{Computed: true, Description: "Start of validity."},
			"valid_to_timestamp":   schema.StringAttribute{Computed: true, Description: "End of validity."},
		},
	}

	resp.Schema = schema.Schema{
		Description: "Manages a route on a secure ingress cluster: a hostname proxied to a VM, Kubernetes cluster or fixed address, with BunkerWeb settings, extra locations, WAF exceptions and an optional custom certificate. Writes fail until the ingress cluster has been deployed.",
		Attributes:  attrs,
	}
}

func (r *SecureIngressRouteResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*apiclient.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *apiclient.Client, got: %T", req.ProviderData))
		return
	}
	r.client = client
}

// --- Normalisation -------------------------------------------------------------

// normalizeSetting reduces a value to a canonical form so that case, list
// order, duplicates and separators do not count as a difference.
func normalizeSetting(s string) string {
	s = strings.TrimSuffix(strings.ToLower(strings.TrimSpace(s)), ".")
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == ',' || r == ';' || r == '|' || r == '\n' || r == '\t'
	})
	sort.Strings(fields)
	out := fields[:0]
	for i, f := range fields {
		if i == 0 || f != fields[i-1] {
			out = append(out, f)
		}
	}
	return strings.Join(out, " ")
}

// keepEquivalent keeps prior when the API returns an equivalent value.
func keepEquivalent(prior types.String, value string) types.String {
	if known(prior) && normalizeSetting(prior.ValueString()) == normalizeSetting(value) {
		return prior
	}
	return types.StringValue(value)
}

// mergeSettings builds the settings state from the API response. Configured
// keys the API omits (it drops values equal to the default) or normalises keep
// their configured value; keys only the API has are added so drift shows up.
func mergeSettings(prior types.Map, api settingsMap) types.Map {
	priorM := mapToStrings(prior)
	if priorM == nil && len(api) == 0 {
		return types.MapNull(types.StringType)
	}
	out := make(map[string]string, len(priorM)+len(api))
	for k, v := range priorM {
		if av, ok := api[k]; ok {
			out[k] = keepEquivalent(types.StringValue(v), av).ValueString()
		} else {
			out[k] = v
		}
	}
	for k, v := range api {
		if _, ok := priorM[k]; !ok {
			out[k] = v
		}
	}
	return stringMap(out)
}

// --- Conversion --------------------------------------------------------------

func routeLocationsBody(plan, state []routeLocationModel) *[]routeLocationBody {
	if plan == nil && len(state) == 0 {
		return nil
	}
	existing := make(map[string]string, len(state))
	for _, l := range state {
		if known(l.Id) {
			existing[l.Path.ValueString()] = l.Id.ValueString()
		}
	}
	out := make([]routeLocationBody, 0, len(plan))
	for _, l := range plan {
		item := routeLocationBody{
			Path:              l.Path.ValueString(),
			TargetKind:        stringPtr(l.TargetKind),
			Virtualmachine:    stringPtr(l.Virtualmachine),
			Interface:         stringPtr(l.Interface),
			KubernetesCluster: stringPtr(l.KubernetesCluster),
			UpstreamAddress:   stringPtr(l.UpstreamAddress),
			Scheme:            l.Scheme.ValueString(),
			Port:              l.Port.ValueInt64(),
			Settings:          mapToStrings(l.Settings),
		}
		if item.Settings == nil {
			item.Settings = map[string]string{}
		}
		if id, ok := existing[item.Path]; ok {
			item.ID = &id
		}
		out = append(out, item)
	}
	return &out
}

func routeWafBodies(plan, state []routeWafModel) *[]routeWafBody {
	if plan == nil && len(state) == 0 {
		return nil
	}
	out := make([]routeWafBody, 0, len(plan))
	for i, w := range plan {
		item := routeWafBody{
			Kind:      w.Kind.ValueString(),
			Rules:     stringPtr(w.Rules),
			Path:      stringPtr(w.Path),
			PathMatch: stringPtr(w.PathMatch),
			Target:    stringPtr(w.Target),
			Raw:       stringPtr(w.Raw),
			Reason:    w.Reason.ValueString(),
		}
		if i < len(state) && known(state[i].Id) {
			id := state[i].Id.ValueInt64()
			item.ID = &id
		}
		out = append(out, item)
	}
	return &out
}

// routeSettingsBody sends configured settings and secrets, and null for keys
// that were removed from the configuration (the API merges settings).
func routeSettingsBody(plan, state *SecureIngressRouteResourceModel) map[string]*string {
	out := map[string]*string{}
	add := func(m map[string]string) {
		for k, v := range m {
			v := v
			out[k] = &v
		}
	}
	planSettings, planSecrets := mapToStrings(plan.Settings), mapToStrings(plan.SecretSettings)
	add(planSettings)
	add(planSecrets)
	if state != nil {
		for _, m := range []map[string]string{mapToStrings(state.Settings), mapToStrings(state.SecretSettings)} {
			for k := range m {
				_, inSettings := planSettings[k]
				_, inSecrets := planSecrets[k]
				if !inSettings && !inSecrets {
					out[k] = nil
				}
			}
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func routeBodyFromPlan(plan, state *SecureIngressRouteResourceModel) routeBody {
	body := routeBody{
		ServerName:        plan.ServerName.ValueString(),
		TargetKind:        stringPtr(plan.TargetKind),
		Virtualmachine:    stringPtr(plan.Virtualmachine),
		Interface:         stringPtr(plan.Interface),
		KubernetesCluster: stringPtr(plan.KubernetesCluster),
		UpstreamAddress:   stringPtr(plan.UpstreamAddress),
		Port:              plan.Port.ValueInt64(),
		Scheme:            plan.Scheme.ValueString(),
		Path:              plan.Path.ValueString(),
		Ssl:               plan.Ssl.ValueBool(),
		Settings:          routeSettingsBody(plan, state),
	}
	var stateLocs []routeLocationModel
	var stateWaf []routeWafModel
	if state == nil {
		body.Cluster = stringPtr(plan.Cluster)
	} else {
		stateLocs, stateWaf = state.Locations, state.WafExceptions
	}
	body.Locations = routeLocationsBody(plan.Locations, stateLocs)
	body.WafExceptions = routeWafBodies(plan.WafExceptions, stateWaf)
	return body
}

func locationFromAPI(a *routeLocationAPI, prior *routeLocationModel) routeLocationModel {
	m := routeLocationModel{Settings: types.MapNull(types.StringType)}
	if prior != nil {
		m = *prior
	}
	m.Id = types.StringValue(a.ID)
	m.Slot = types.Int64Value(a.Slot)
	m.Path = types.StringValue(a.Path)
	m.TargetKind = types.StringValue(a.TargetKind)
	m.Virtualmachine = strOrNull(a.Virtualmachine)
	m.Interface = strOrNull(a.Interface)
	m.KubernetesCluster = strOrNull(a.KubernetesCluster)
	m.UpstreamAddress = keepEquivalent(m.UpstreamAddress, a.UpstreamAddress)
	m.Scheme = types.StringValue(a.Scheme)
	m.Port = types.Int64Value(a.Port)
	m.Settings = mergeSettings(m.Settings, a.Settings)
	m.Upstream = types.StringValue(a.Upstream)
	return m
}

func wafFromAPI(a *routeWafAPIModel, prior *routeWafModel) routeWafModel {
	var p routeWafModel
	if prior != nil {
		p = *prior
	}
	return routeWafModel{
		Id:        types.Int64Value(a.ID),
		Kind:      types.StringValue(a.Kind),
		Rules:     keepEquivalent(p.Rules, a.Rules),
		Path:      keepEquivalent(p.Path, a.Path),
		PathMatch: types.StringValue(a.PathMatch),
		Target:    keepEquivalent(p.Target, a.Target),
		Raw:       keepEquivalent(p.Raw, a.Raw),
		Reason:    keepEquivalent(p.Reason, a.Reason),
	}
}

func certFromAPI(a *routeCertAPIModel, prior *routeCertModel) *routeCertModel {
	if a == nil {
		return nil
	}
	m := routeCertModel{
		Format:      types.StringNull(),
		Certificate: types.StringNull(),
		PrivateKey:  types.StringNull(),
		Passphrase:  types.StringNull(),
	}
	if prior != nil {
		m.Format, m.Certificate, m.PrivateKey, m.Passphrase = prior.Format, prior.Certificate, prior.PrivateKey, prior.Passphrase
	}
	m.Id = types.StringValue(a.ID)
	m.CommonName = types.StringValue(a.CommonName)
	m.SubjectAltNames = stringList(a.SubjectAltNames)
	m.Issuer = types.StringValue(a.Issuer)
	m.SerialNumber = types.StringValue(a.SerialNumber)
	m.FingerprintSha256 = types.StringValue(a.FingerprintSha256)
	m.ValidFromTimestamp = strOrNull(a.ValidFromTimestamp)
	m.ValidToTimestamp = strOrNull(a.ValidToTimestamp)
	return &m
}

// routeStateFromAPI updates m (the plan or prior state) from the API response.
func routeStateFromAPI(a *routeAPIModel, m *SecureIngressRouteResourceModel) {
	m.Id = types.StringValue(a.ID)
	m.Cluster = types.StringValue(a.Cluster)
	m.ServerName = keepEquivalent(m.ServerName, a.ServerName)
	m.TargetKind = types.StringValue(a.TargetKind)
	m.Virtualmachine = strOrNull(a.Virtualmachine)
	m.Interface = strOrNull(a.Interface)
	m.KubernetesCluster = strOrNull(a.KubernetesCluster)
	m.UpstreamAddress = keepEquivalent(m.UpstreamAddress, a.UpstreamAddress)
	m.Port = types.Int64Value(a.Port)
	m.Scheme = types.StringValue(a.Scheme)
	m.Path = types.StringValue(a.Path)
	m.Ssl = types.BoolValue(a.Ssl)
	m.Settings = mergeSettings(m.Settings, a.Settings)
	m.SecretsSet = stringList(a.SecretsSet)
	m.Upstream = types.StringValue(a.Upstream)

	// Locations: keep the configured order, matching by path.
	byPath := make(map[string]*routeLocationAPI, len(a.Locations))
	for i := range a.Locations {
		byPath[a.Locations[i].Path] = &a.Locations[i]
	}
	var locs []routeLocationModel
	used := map[string]bool{}
	for i := range m.Locations {
		if api, ok := byPath[m.Locations[i].Path.ValueString()]; ok {
			locs = append(locs, locationFromAPI(api, &m.Locations[i]))
			used[api.Path] = true
		}
	}
	for i := range a.Locations {
		if !used[a.Locations[i].Path] {
			locs = append(locs, locationFromAPI(&a.Locations[i], nil))
		}
	}
	if locs == nil && m.Locations != nil {
		locs = []routeLocationModel{}
	}
	m.Locations = locs

	var waf []routeWafModel
	for i := range a.WafExceptions {
		var prior *routeWafModel
		if i < len(m.WafExceptions) {
			prior = &m.WafExceptions[i]
		}
		waf = append(waf, wafFromAPI(&a.WafExceptions[i], prior))
	}
	if waf == nil && m.WafExceptions != nil {
		waf = []routeWafModel{}
	}
	m.WafExceptions = waf

	m.CustomCertificate = certFromAPI(a.CustomCertificate, m.CustomCertificate)
}

// --- Errors -------------------------------------------------------------------

func routeError(action string, err error) (string, string) {
	if apiclient.IsConflict(err) {
		return "Ingress cluster not deployed yet",
			fmt.Sprintf("Could not %s the secure ingress route because its ingress cluster has not been deployed yet. "+
				"Wait until the cluster has been synced, then run terraform apply again.\n\nAPI response: %s", action, err)
	}
	return "Error trying to " + action + " secure ingress route", err.Error()
}

// --- Certificate ---------------------------------------------------------------

func certInputsEqual(a, b *routeCertModel) bool {
	if a == nil || b == nil {
		return a == b
	}
	return a.Format.Equal(b.Format) && a.Certificate.Equal(b.Certificate) &&
		a.PrivateKey.Equal(b.PrivateKey) && a.Passphrase.Equal(b.Passphrase)
}

// certificateBytes returns the file content to upload. Binary bundles are
// configured as base64 (e.g. with filebase64()).
func certificateBytes(format, content string) []byte {
	if format != "PEM" && !strings.HasPrefix(strings.TrimSpace(content), "-----BEGIN") {
		if decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(content)); err == nil {
			return decoded
		}
	}
	return []byte(content)
}

func (r *SecureIngressRouteResource) uploadCertificate(ctx context.Context, id string, c *routeCertModel) error {
	format := c.Format.ValueString()
	fields := map[string]string{"format": format}
	if known(c.Passphrase) {
		fields["passphrase"] = c.Passphrase.ValueString()
	}
	files := map[string][]byte{"certificate": certificateBytes(format, c.Certificate.ValueString())}
	if known(c.PrivateKey) {
		files["privatekey"] = []byte(c.PrivateKey.ValueString())
	}
	return r.client.PostMultipart(ctx, routeCollection+id+"/certificate/", fields, files, nil)
}

// --- CRUD -------------------------------------------------------------------------

func (r *SecureIngressRouteResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecureIngressRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result routeAPIModel
	if err := r.client.Post(ctx, routeCollection, routeBodyFromPlan(&plan, nil), &result); err != nil {
		resp.Diagnostics.AddError(routeError("create", err))
		return
	}

	if cert := plan.CustomCertificate; cert != nil {
		// Save the route first so a failed upload does not orphan it.
		routeStateFromAPI(&result, &plan)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		if err := r.uploadCertificate(ctx, result.ID, cert); err != nil {
			resp.Diagnostics.AddError(routeError("upload the custom certificate for", err))
			return
		}
		plan.CustomCertificate = cert
		if err := r.client.Get(ctx, routeCollection+result.ID+"/", &result); err != nil {
			resp.Diagnostics.AddError(routeError("read", err))
			return
		}
	}

	routeStateFromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SecureIngressRouteResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecureIngressRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var result routeAPIModel
	if err := r.client.Get(ctx, routeCollection+state.Id.ValueString()+"/", &result); err != nil {
		if apiclient.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError(routeError("read", err))
		return
	}

	routeStateFromAPI(&result, &state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *SecureIngressRouteResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state SecureIngressRouteResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Id.ValueString()

	// Revert to Let's Encrypt before changing server_name, since the API
	// rejects a hostname the current certificate does not cover.
	certRemoved := plan.CustomCertificate == nil && state.CustomCertificate != nil
	if certRemoved {
		if err := r.client.Delete(ctx, routeCollection+id+"/certificate/"); err != nil && !apiclient.IsNotFound(err) {
			resp.Diagnostics.AddError(routeError("remove the custom certificate from", err))
			return
		}
	}

	var result routeAPIModel
	if err := r.client.Patch(ctx, routeCollection+id+"/", routeBodyFromPlan(&plan, &state), &result); err != nil {
		resp.Diagnostics.AddError(routeError("update", err))
		return
	}

	if plan.CustomCertificate != nil && !certInputsEqual(plan.CustomCertificate, state.CustomCertificate) {
		if err := r.uploadCertificate(ctx, id, plan.CustomCertificate); err != nil {
			resp.Diagnostics.AddError(routeError("upload the custom certificate for", err))
			return
		}
		if err := r.client.Get(ctx, routeCollection+id+"/", &result); err != nil {
			resp.Diagnostics.AddError(routeError("read", err))
			return
		}
	}

	routeStateFromAPI(&result, &plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *SecureIngressRouteResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecureIngressRouteResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Delete(ctx, routeCollection+state.Id.ValueString()+"/")
	if err != nil && !apiclient.IsNotFound(err) {
		resp.Diagnostics.AddError(routeError("delete", err))
	}
}

func (r *SecureIngressRouteResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
