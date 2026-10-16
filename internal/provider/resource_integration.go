package provider

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IntegrationResourceModel struct {
	Id           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	TypeId       types.Int64  `tfsdk:"type_id"`
	TypeName     types.String `tfsdk:"type_name"`
	CatalogKey   types.String `tfsdk:"catalog_key"`
	Subtype      types.String `tfsdk:"subtype"`
	Baseurl      types.String `tfsdk:"baseurl"`
	Username     types.String `tfsdk:"username"`
	Secret       types.String `tfsdk:"secret"`
	Key          types.String `tfsdk:"key"`
	SecretExpire types.String `tfsdk:"secret_expire"`
	Icon         types.String `tfsdk:"icon"`
	HasSecret    types.Bool   `tfsdk:"has_secret"`
	HasKey       types.Bool   `tfsdk:"has_key"`
}

// flexBool decodes a JSON boolean or a string such as "true"/"True".
type flexBool bool

func (b *flexBool) UnmarshalJSON(data []byte) error {
	var v bool
	if err := json.Unmarshal(data, &v); err == nil {
		*b = flexBool(v)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		*b = false
		return nil
	}
	*b = flexBool(strings.EqualFold(s, "true"))
	return nil
}

type integrationAPIModel struct {
	ID           int64                 `json:"id"`
	Name         string                `json:"name"`
	Type         struct{ Name string } `json:"type"`
	Subtype      string                `json:"subtype"`
	Baseurl      string                `json:"baseurl"`
	Username     string                `json:"username"`
	SecretExpire *string               `json:"secret_expire"`
	Icon         string                `json:"icon"`
	HasSecret    flexBool              `json:"has_secret"`
	HasKey       flexBool              `json:"has_key"`
}

type integrationBody struct {
	Name         string  `json:"name"`
	TypeId       *int64  `json:"type_id,omitempty"`
	CatalogKey   *string `json:"catalog_key,omitempty"`
	Subtype      *string `json:"subtype,omitempty"`
	Baseurl      string  `json:"baseurl"`
	Username     *string `json:"username,omitempty"`
	Secret       *string `json:"secret,omitempty"`
	Key          *string `json:"key,omitempty"`
	SecretExpire *string `json:"secret_expire,omitempty"`
}

// sameInstant keeps the configured timestamp when the API returns the same
// moment in a different format.
func sameInstant(prior types.String, value *string) types.String {
	if value == nil {
		return types.StringNull()
	}
	if known(prior) {
		a, errA := time.Parse(time.RFC3339, prior.ValueString())
		b, errB := time.Parse(time.RFC3339, *value)
		if errA == nil && errB == nil && a.Equal(b) {
			return prior
		}
	}
	return types.StringValue(*value)
}

func sensitiveWriteOnly(desc string) schema.StringAttribute {
	return schema.StringAttribute{Optional: true, Sensitive: true, Description: desc}
}

func NewIntegrationResource() resource.Resource {
	return &crudResource[IntegrationResourceModel, integrationAPIModel]{
		typeSuffix: "integration",
		label:      "integration",
		collection: "/api/tenant/integrations/",
		schema: schema.Schema{
			Description: "Manages an integration with an external system within the tenant. Credentials (`secret`, `key`) are write-only: the API never returns them, so Terraform keeps the configured values and cannot detect changes made outside Terraform.",
			Attributes: map[string]schema.Attribute{
				"id":   rsIntID(),
				"name": rsRequiredString("Name of the integration. Cannot contain special characters or a dot."),
				"type_id": schema.Int64Attribute{
					Optional:    true,
					Description: "ID of the integration type. Alternatively set `catalog_key` to resolve the type from the integration catalog.",
				},
				"type_name": rsVolatileString("Name of the integration type."),
				"catalog_key": schema.StringAttribute{
					Optional:      true,
					Description:   "Integration catalog entry used to resolve the type, subtype and icon on creation. Changing it forces a new integration.",
					PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				},
				"subtype":       rsOptComputedString("Integration subtype. Resolved from `catalog_key` when omitted."),
				"baseurl":       rsRequiredString("Base URL of the external system."),
				"username":      rsOptComputedString("Username used to authenticate."),
				"secret":        sensitiveWriteOnly("Secret or password used to authenticate. Write-only."),
				"key":           sensitiveWriteOnly("API key used to authenticate. Write-only."),
				"secret_expire": rsOptComputedString("Expiry timestamp of the credentials (RFC 3339)."),
				"icon":          rsVolatileString("Icon of the integration."),
				"has_secret":    rsVolatileBool("Whether a secret is stored."),
				"has_key":       rsVolatileBool("Whether a key is stored."),
			},
		},
		idOf: func(m *IntegrationResourceModel) string { return strconv.FormatInt(m.Id.ValueInt64(), 10) },
		toAPI: func(m *IntegrationResourceModel) any {
			return integrationBody{
				Name:         m.Name.ValueString(),
				TypeId:       int64Ptr(m.TypeId),
				CatalogKey:   stringPtr(m.CatalogKey),
				Subtype:      stringPtr(m.Subtype),
				Baseurl:      m.Baseurl.ValueString(),
				Username:     stringPtr(m.Username),
				Secret:       stringPtr(m.Secret),
				Key:          stringPtr(m.Key),
				SecretExpire: stringPtr(m.SecretExpire),
			}
		},
		fromAPI: func(a *integrationAPIModel, m *IntegrationResourceModel) {
			m.Id = types.Int64Value(a.ID)
			m.Name = types.StringValue(a.Name)
			m.TypeName = types.StringValue(a.Type.Name)
			m.Subtype = types.StringValue(a.Subtype)
			m.Baseurl = types.StringValue(a.Baseurl)
			m.Username = types.StringValue(a.Username)
			m.SecretExpire = sameInstant(m.SecretExpire, a.SecretExpire)
			m.Icon = types.StringValue(a.Icon)
			m.HasSecret = types.BoolValue(bool(a.HasSecret))
			m.HasKey = types.BoolValue(bool(a.HasKey))
		},
	}
}
