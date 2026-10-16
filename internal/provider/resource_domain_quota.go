package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type DomainQuotaResourceModel struct {
	Id               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Description      types.String `tfsdk:"description"`
	MaxRules         types.Int64  `tfsdk:"max_rules"`
	MaxObjects       types.Int64  `tfsdk:"max_objects"`
	MaxObjectgroups  types.Int64  `tfsdk:"max_objectgroups"`
	MaxServices      types.Int64  `tfsdk:"max_services"`
	MaxServicegroups types.Int64  `tfsdk:"max_servicegroups"`
	MaxNetworks      types.Int64  `tfsdk:"max_networks"`
}

type domainQuotaAPIModel struct {
	ID               string  `json:"id,omitempty"`
	Name             string  `json:"name"`
	Description      *string `json:"description,omitempty"`
	MaxRules         *int64  `json:"max_rules,omitempty"`
	MaxObjects       *int64  `json:"max_objects,omitempty"`
	MaxObjectgroups  *int64  `json:"max_objectgroups,omitempty"`
	MaxServices      *int64  `json:"max_services,omitempty"`
	MaxServicegroups *int64  `json:"max_servicegroups,omitempty"`
	MaxNetworks      *int64  `json:"max_networks,omitempty"`
}

func int64Value(i *int64) types.Int64 {
	if i == nil {
		return types.Int64Value(0)
	}
	return types.Int64Value(*i)
}

func NewDomainQuotaResource() resource.Resource {
	return &crudResource[DomainQuotaResourceModel, domainQuotaAPIModel]{
		typeSuffix: "domain_quota",
		label:      "domain quota",
		collection: "/api/tenant/domainquotas/",
		schema: schema.Schema{
			Description: "Manages a domain quota profile, limiting the number of firewall objects in a domain. Quota profiles are global (not tenant-scoped).",
			Attributes: map[string]schema.Attribute{
				"id":                rsID(),
				"name":              rsRequiredString("Name of the quota profile."),
				"description":       rsOptComputedString("Description of the quota profile."),
				"max_rules":         rsOptComputedInt64("Max number of firewall rules in a domain."),
				"max_objects":       rsOptComputedInt64("Max number of firewall objects in a domain."),
				"max_objectgroups":  rsOptComputedInt64("Max number of firewall object groups in a domain."),
				"max_services":      rsOptComputedInt64("Max number of firewall services in a domain."),
				"max_servicegroups": rsOptComputedInt64("Max number of firewall service groups in a domain."),
				"max_networks":      rsOptComputedInt64("Max number of networks (VLANs) in a domain."),
			},
		},
		idOf: func(m *DomainQuotaResourceModel) string { return m.Id.ValueString() },
		toAPI: func(m *DomainQuotaResourceModel) any {
			return domainQuotaAPIModel{
				Name:             m.Name.ValueString(),
				Description:      stringPtr(m.Description),
				MaxRules:         int64Ptr(m.MaxRules),
				MaxObjects:       int64Ptr(m.MaxObjects),
				MaxObjectgroups:  int64Ptr(m.MaxObjectgroups),
				MaxServices:      int64Ptr(m.MaxServices),
				MaxServicegroups: int64Ptr(m.MaxServicegroups),
				MaxNetworks:      int64Ptr(m.MaxNetworks),
			}
		},
		fromAPI: func(a *domainQuotaAPIModel, m *DomainQuotaResourceModel) {
			m.Id = types.StringValue(a.ID)
			m.Name = types.StringValue(a.Name)
			m.Description = types.StringValue(derefString(a.Description))
			m.MaxRules = int64Value(a.MaxRules)
			m.MaxObjects = int64Value(a.MaxObjects)
			m.MaxObjectgroups = int64Value(a.MaxObjectgroups)
			m.MaxServices = int64Value(a.MaxServices)
			m.MaxServicegroups = int64Value(a.MaxServicegroups)
			m.MaxNetworks = int64Value(a.MaxNetworks)
		},
	}
}
