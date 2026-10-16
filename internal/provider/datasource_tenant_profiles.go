package provider

import (
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type domainQuotaItem struct {
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

func NewDomainQuotaDataSource() datasource.DataSource {
	return &listDataSource[domainQuotaAPIModel, domainQuotaItem]{
		typeSuffix:  "domain_quota",
		label:       "domain quota",
		description: "Looks up domain quota profiles by ID or name, or lists all of them.",
		collection:  "/api/tenant/domainquotas/",
		pluralAttr:  "domain_quotas",
		nameFilter:  "name__icontains",
		itemAttrs: map[string]dschema.Attribute{
			"id":                dsString("Quota profile ID."),
			"name":              dsString("Name of the quota profile."),
			"description":       dsString("Description of the quota profile."),
			"max_rules":         dsInt64("Max number of firewall rules in a domain."),
			"max_objects":       dsInt64("Max number of firewall objects in a domain."),
			"max_objectgroups":  dsInt64("Max number of firewall object groups in a domain."),
			"max_services":      dsInt64("Max number of firewall services in a domain."),
			"max_servicegroups": dsInt64("Max number of firewall service groups in a domain."),
			"max_networks":      dsInt64("Max number of networks (VLANs) in a domain."),
		},
		nameOf: func(a *domainQuotaAPIModel) string { return a.Name },
		toItem: func(a *domainQuotaAPIModel) domainQuotaItem {
			return domainQuotaItem{
				Id:               types.StringValue(a.ID),
				Name:             types.StringValue(a.Name),
				Description:      types.StringValue(derefString(a.Description)),
				MaxRules:         int64Value(a.MaxRules),
				MaxObjects:       int64Value(a.MaxObjects),
				MaxObjectgroups:  int64Value(a.MaxObjectgroups),
				MaxServices:      int64Value(a.MaxServices),
				MaxServicegroups: int64Value(a.MaxServicegroups),
				MaxNetworks:      int64Value(a.MaxNetworks),
			}
		},
	}
}

type loadbalancingQuotaItem struct {
	Id                     types.String `tfsdk:"id"`
	Name                   types.String `tfsdk:"name"`
	Description            types.String `tfsdk:"description"`
	MaxPoliciesInCsvserver types.Int64  `tfsdk:"max_policies_in_csvserver"`
}

func NewLoadbalancingQuotaDataSource() datasource.DataSource {
	return &listDataSource[loadbalancingQuotaAPIModel, loadbalancingQuotaItem]{
		typeSuffix:  "loadbalancing_quota",
		label:       "loadbalancing quota",
		description: "Looks up loadbalancing quota profiles by ID or name, or lists all of them.",
		collection:  "/api/tenant/loadbalancingquotas/",
		pluralAttr:  "loadbalancing_quotas",
		nameFilter:  "name__icontains",
		itemAttrs: map[string]dschema.Attribute{
			"id":                        dsString("Quota profile ID."),
			"name":                      dsString("Name of the quota profile."),
			"description":               dsString("Description of the quota profile."),
			"max_policies_in_csvserver": dsInt64("Max number of policies linked to a content switching virtual server."),
		},
		nameOf: func(a *loadbalancingQuotaAPIModel) string { return a.Name },
		toItem: func(a *loadbalancingQuotaAPIModel) loadbalancingQuotaItem {
			return loadbalancingQuotaItem{
				Id:                     types.StringValue(a.ID),
				Name:                   types.StringValue(a.Name),
				Description:            types.StringValue(derefString(a.Description)),
				MaxPoliciesInCsvserver: int64Value(a.MaxPoliciesInCsvserver),
			}
		},
	}
}

type quotaItem struct {
	Id                 types.String `tfsdk:"id"`
	Name               types.String `tfsdk:"name"`
	Description        types.String `tfsdk:"description"`
	DomainQuota        types.String `tfsdk:"domain_quota"`
	LoadbalancingQuota types.String `tfsdk:"loadbalancing_quota"`
}

func NewQuotaDataSource() datasource.DataSource {
	return &listDataSource[quotaAPIModel, quotaItem]{
		typeSuffix:  "quota",
		label:       "quota",
		description: "Looks up quota profiles by ID or name, or lists all of them.",
		collection:  "/api/tenant/quotas/",
		pluralAttr:  "quotas",
		nameFilter:  "name__icontains",
		itemAttrs: map[string]dschema.Attribute{
			"id":                  dsString("Quota profile ID."),
			"name":                dsString("Name of the quota profile."),
			"description":         dsString("Description of the quota profile."),
			"domain_quota":        dsString("ID of the domain quota profile."),
			"loadbalancing_quota": dsString("ID of the loadbalancing quota profile."),
		},
		nameOf: func(a *quotaAPIModel) string { return a.Name },
		toItem: func(a *quotaAPIModel) quotaItem {
			return quotaItem{
				Id:                 types.StringValue(a.ID),
				Name:               types.StringValue(a.Name),
				Description:        types.StringValue(derefString(a.Description)),
				DomainQuota:        types.StringValue(derefString(a.DomainQuota)),
				LoadbalancingQuota: types.StringValue(derefString(a.LoadbalancingQuota)),
			}
		},
	}
}

type integrationTypeItem struct {
	Id   types.Int64  `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func NewIntegrationTypeDataSource() datasource.DataSource {
	return &listDataSource[integrationTypeAPIModel, integrationTypeItem]{
		typeSuffix:  "integration_type",
		label:       "integration type",
		description: "Looks up integration types by ID or name, or lists all of them.",
		collection:  "/api/tenant/integrationtypes/",
		pluralAttr:  "integration_types",
		nameFilter:  "name__icontains",
		idIsInt:     true,
		itemAttrs: map[string]dschema.Attribute{
			"id":   dsInt64("Integration type ID."),
			"name": dsString("Name of the integration type."),
		},
		nameOf: func(a *integrationTypeAPIModel) string { return a.Name },
		toItem: func(a *integrationTypeAPIModel) integrationTypeItem {
			return integrationTypeItem{Id: types.Int64Value(a.ID), Name: types.StringValue(a.Name)}
		},
	}
}

type integrationItem struct {
	Id           types.Int64  `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	TypeName     types.String `tfsdk:"type_name"`
	Subtype      types.String `tfsdk:"subtype"`
	Baseurl      types.String `tfsdk:"baseurl"`
	Username     types.String `tfsdk:"username"`
	SecretExpire types.String `tfsdk:"secret_expire"`
	Icon         types.String `tfsdk:"icon"`
	HasSecret    types.Bool   `tfsdk:"has_secret"`
	HasKey       types.Bool   `tfsdk:"has_key"`
}

func NewIntegrationDataSource() datasource.DataSource {
	return &listDataSource[integrationAPIModel, integrationItem]{
		typeSuffix:  "integration",
		label:       "integration",
		description: "Looks up integrations by ID or name, or lists all of them. Credentials are never returned.",
		collection:  "/api/tenant/integrations/",
		pluralAttr:  "integrations",
		nameFilter:  "name__icontains",
		idIsInt:     true,
		itemAttrs: map[string]dschema.Attribute{
			"id":            dsInt64("Integration ID."),
			"name":          dsString("Name of the integration."),
			"type_name":     dsString("Name of the integration type."),
			"subtype":       dsString("Integration subtype."),
			"baseurl":       dsString("Base URL of the external system."),
			"username":      dsString("Username used to authenticate."),
			"secret_expire": dsString("Expiry timestamp of the credentials."),
			"icon":          dsString("Icon of the integration."),
			"has_secret":    dsBool("Whether a secret is stored."),
			"has_key":       dsBool("Whether a key is stored."),
		},
		nameOf: func(a *integrationAPIModel) string { return a.Name },
		toItem: func(a *integrationAPIModel) integrationItem {
			return integrationItem{
				Id:           types.Int64Value(a.ID),
				Name:         types.StringValue(a.Name),
				TypeName:     types.StringValue(a.Type.Name),
				Subtype:      types.StringValue(a.Subtype),
				Baseurl:      types.StringValue(a.Baseurl),
				Username:     types.StringValue(a.Username),
				SecretExpire: strOrNull(a.SecretExpire),
				Icon:         types.StringValue(a.Icon),
				HasSecret:    types.BoolValue(bool(a.HasSecret)),
				HasKey:       types.BoolValue(bool(a.HasKey)),
			}
		},
	}
}
