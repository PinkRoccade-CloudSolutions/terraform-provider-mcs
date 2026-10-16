package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type QuotaResourceModel struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	DomainQuota        types.String `tfsdk:"domain_quota"`
	LoadbalancingQuota types.String `tfsdk:"loadbalancing_quota"`
}

type quotaAPIModel struct {
	ID                 string  `json:"id,omitempty"`
	Name               string  `json:"name"`
	Description        *string `json:"description,omitempty"`
	DomainQuota        *string `json:"domain_quota,omitempty"`
	LoadbalancingQuota *string `json:"loadbalancing_quota,omitempty"`
}

func NewQuotaResource() resource.Resource {
	return &crudResource[QuotaResourceModel, quotaAPIModel]{
		typeSuffix: "quota",
		label:      "quota",
		collection: "/api/tenant/quotas/",
		schema: schema.Schema{
			Description: "Manages a quota profile combining a domain quota and a loadbalancing quota. Quota profiles are global (not tenant-scoped).",
			Attributes: map[string]schema.Attribute{
				"id":                  rsID(),
				"name":                rsRequiredString("Name of the quota profile."),
				"description":         rsOptComputedString("Description of the quota profile."),
				"domain_quota":        rsOptComputedString("ID of the domain quota profile."),
				"loadbalancing_quota": rsOptComputedString("ID of the loadbalancing quota profile."),
			},
		},
		idOf: func(m *QuotaResourceModel) string { return m.Id.ValueString() },
		toAPI: func(m *QuotaResourceModel) any {
			return quotaAPIModel{
				Name:               m.Name.ValueString(),
				Description:        stringPtr(m.Description),
				DomainQuota:        stringPtr(m.DomainQuota),
				LoadbalancingQuota: stringPtr(m.LoadbalancingQuota),
			}
		},
		fromAPI: func(a *quotaAPIModel, m *QuotaResourceModel) {
			m.Id = types.StringValue(a.ID)
			m.Name = types.StringValue(a.Name)
			m.Description = types.StringValue(derefString(a.Description))
			m.DomainQuota = types.StringValue(derefString(a.DomainQuota))
			m.LoadbalancingQuota = types.StringValue(derefString(a.LoadbalancingQuota))
		},
	}
}
