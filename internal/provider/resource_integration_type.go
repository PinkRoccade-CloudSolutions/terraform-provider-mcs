package provider

import (
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type IntegrationTypeResourceModel struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

type integrationTypeAPIModel struct {
	ID   int64  `json:"id,omitempty"`
	Name string `json:"name"`
}

func NewIntegrationTypeResource() resource.Resource {
	return &crudResource[IntegrationTypeResourceModel, integrationTypeAPIModel]{
		typeSuffix: "integration_type",
		label:      "integration type",
		collection: "/api/tenant/integrationtypes/",
		schema: schema.Schema{
			Description: "Manages an integration type within the tenant.",
			Attributes: map[string]schema.Attribute{
				"id":   rsIntID(),
				"name": rsRequiredString("Name of the integration type."),
			},
		},
		idOf: func(m *IntegrationTypeResourceModel) string { return strconv.FormatInt(m.Id.ValueInt64(), 10) },
		toAPI: func(m *IntegrationTypeResourceModel) any {
			return integrationTypeAPIModel{Name: m.Name.ValueString()}
		},
		fromAPI: func(a *integrationTypeAPIModel, m *IntegrationTypeResourceModel) {
			m.Id = types.Int64Value(a.ID)
			m.Name = types.StringValue(a.Name)
		},
	}
}
