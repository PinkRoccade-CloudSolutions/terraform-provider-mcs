package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type LoadbalancingQuotaResourceModel struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	MaxPoliciesInCsvserver types.Int64  `tfsdk:"max_policies_in_csvserver"`
}

type loadbalancingQuotaAPIModel struct {
	ID                     string  `json:"id,omitempty"`
	Name                   string  `json:"name"`
	Description            *string `json:"description,omitempty"`
	MaxPoliciesInCsvserver *int64  `json:"max_policies_in_csvserver,omitempty"`
}

func NewLoadbalancingQuotaResource() resource.Resource {
	return &crudResource[LoadbalancingQuotaResourceModel, loadbalancingQuotaAPIModel]{
		typeSuffix: "loadbalancing_quota",
		label:      "loadbalancing quota",
		collection: "/api/tenant/loadbalancingquotas/",
		schema: schema.Schema{
			Description: "Manages a loadbalancing quota profile. Quota profiles are global (not tenant-scoped).",
			Attributes: map[string]schema.Attribute{
				"id":                        rsID(),
				"name":                      rsRequiredString("Name of the quota profile."),
				"description":               rsOptComputedString("Description of the quota profile."),
				"max_policies_in_csvserver": rsOptComputedInt64("Max number of policies linked to a content switching virtual server."),
			},
		},
		idOf: func(m *LoadbalancingQuotaResourceModel) string { return m.Id.ValueString() },
		toAPI: func(m *LoadbalancingQuotaResourceModel) any {
			return loadbalancingQuotaAPIModel{
				Name:                   m.Name.ValueString(),
				Description:            stringPtr(m.Description),
				MaxPoliciesInCsvserver: int64Ptr(m.MaxPoliciesInCsvserver),
			}
		},
		fromAPI: func(a *loadbalancingQuotaAPIModel, m *LoadbalancingQuotaResourceModel) {
			m.Id = types.StringValue(a.ID)
			m.Name = types.StringValue(a.Name)
			m.Description = types.StringValue(derefString(a.Description))
			m.MaxPoliciesInCsvserver = int64Value(a.MaxPoliciesInCsvserver)
		},
	}
}
