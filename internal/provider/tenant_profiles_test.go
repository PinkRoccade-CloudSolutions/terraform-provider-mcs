package provider

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// ---------------------------------------------------------------------------
// mcs_loadbalancer_pool
// ---------------------------------------------------------------------------

func TestAccLoadBalancerPoolResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	pool := mock.OnStateful("/api/networking/loadbalancerpools", map[string]interface{}{
		"id": "lbp-001", "description": "", "ip_pool": "ipp-001", "ip_pool_name": "lbpools-1",
		"prefix": "203.0.113.16/28", "status": "free", "customer": "acme",
		"assigned_object_type": "", "assigned_object_id": nil, "assigned_object_label": "",
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_loadbalancer_pool" "test" {
  name    = "web-pool"
  ip_pool = "ipp-001"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_loadbalancer_pool.test", "id", "lbp-001"),
					resource.TestCheckResourceAttr("mcs_loadbalancer_pool.test", "name", "web-pool"),
					resource.TestCheckResourceAttr("mcs_loadbalancer_pool.test", "description", ""),
					resource.TestCheckResourceAttr("mcs_loadbalancer_pool.test", "ip_pool", "ipp-001"),
					resource.TestCheckResourceAttr("mcs_loadbalancer_pool.test", "ip_pool_name", "lbpools-1"),
					resource.TestCheckResourceAttr("mcs_loadbalancer_pool.test", "prefix", "203.0.113.16/28"),
					resource.TestCheckResourceAttr("mcs_loadbalancer_pool.test", "status", "free"),
					resource.TestCheckResourceAttr("mcs_loadbalancer_pool.test", "customer", "acme"),
					resource.TestCheckResourceAttr("mcs_loadbalancer_pool.test", "assigned_object_type", ""),
					resource.TestCheckNoResourceAttr("mcs_loadbalancer_pool.test", "assigned_object_id"),
					resource.TestCheckResourceAttr("mcs_loadbalancer_pool.test", "assigned_object_label", ""),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_loadbalancer_pool" "test" {
  name        = "web-pool-2"
  description = "Web frontends"
  ip_pool     = "ipp-001"
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("mcs_loadbalancer_pool.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_loadbalancer_pool.test", "name", "web-pool-2"),
					resource.TestCheckResourceAttr("mcs_loadbalancer_pool.test", "description", "Web frontends"),
				),
			},
			{
				PreConfig:   func() { pool.SetDeleteStatus(http.StatusConflict) },
				Config:      providerConfigBlock(mock.URL()),
				ExpectError: regexpMustCompile(`pool is still in use`),
			},
			{
				PreConfig: func() { pool.SetDeleteStatus(0) },
				Config:    providerConfigBlock(mock.URL()),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// Tenant quotas and integrations
// ---------------------------------------------------------------------------

func TestAccDomainQuotaResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.OnStateful("/api/tenant/domainquotas", map[string]interface{}{
		"id": "dq-001", "description": "", "max_rules": 500, "max_objects": 1000,
		"max_objectgroups": 100, "max_services": 200, "max_servicegroups": 50, "max_networks": 10,
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_domain_quota" "test" {
  name = "standard"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "id", "dq-001"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "name", "standard"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "description", ""),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "max_rules", "500"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "max_objects", "1000"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "max_objectgroups", "100"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "max_services", "200"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "max_servicegroups", "50"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "max_networks", "10"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_domain_quota" "test" {
  name              = "standard"
  description       = "Standard tier"
  max_rules         = 750
  max_objects       = 1500
  max_objectgroups  = 150
  max_services      = 300
  max_servicegroups = 75
  max_networks      = 20
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "description", "Standard tier"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "max_rules", "750"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "max_objects", "1500"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "max_objectgroups", "150"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "max_services", "300"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "max_servicegroups", "75"),
					resource.TestCheckResourceAttr("mcs_domain_quota.test", "max_networks", "20"),
				),
			},
		},
	})
}

func TestAccLoadbalancingQuotaResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.OnStateful("/api/tenant/loadbalancingquotas", map[string]interface{}{
		"id": "lq-001", "description": "", "max_policies_in_csvserver": 25,
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_loadbalancing_quota" "test" {
  name = "standard"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_loadbalancing_quota.test", "id", "lq-001"),
					resource.TestCheckResourceAttr("mcs_loadbalancing_quota.test", "name", "standard"),
					resource.TestCheckResourceAttr("mcs_loadbalancing_quota.test", "description", ""),
					resource.TestCheckResourceAttr("mcs_loadbalancing_quota.test", "max_policies_in_csvserver", "25"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_loadbalancing_quota" "test" {
  name                      = "standard"
  description               = "Standard tier"
  max_policies_in_csvserver = 40
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_loadbalancing_quota.test", "description", "Standard tier"),
					resource.TestCheckResourceAttr("mcs_loadbalancing_quota.test", "max_policies_in_csvserver", "40"),
				),
			},
		},
	})
}

func TestAccQuotaResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.OnStateful("/api/tenant/quotas", map[string]interface{}{
		"id": "q-001", "description": "", "domain_quota": nil, "loadbalancing_quota": nil,
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_quota" "test" {
  name = "standard"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_quota.test", "id", "q-001"),
					resource.TestCheckResourceAttr("mcs_quota.test", "name", "standard"),
					resource.TestCheckResourceAttr("mcs_quota.test", "description", ""),
					resource.TestCheckResourceAttr("mcs_quota.test", "domain_quota", ""),
					resource.TestCheckResourceAttr("mcs_quota.test", "loadbalancing_quota", ""),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_quota" "test" {
  name                = "standard"
  description         = "Standard tier"
  domain_quota        = "dq-001"
  loadbalancing_quota = "lq-001"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_quota.test", "description", "Standard tier"),
					resource.TestCheckResourceAttr("mcs_quota.test", "domain_quota", "dq-001"),
					resource.TestCheckResourceAttr("mcs_quota.test", "loadbalancing_quota", "lq-001"),
				),
			},
			{
				ResourceName:      "mcs_quota.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIntegrationTypeResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.OnStateful("/api/tenant/integrationtypes", map[string]interface{}{"id": 4})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		CheckDestroy:             mock.checkCalled(http.MethodDelete, "/api/tenant/integrationtypes/4/"),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_integration_type" "test" {
  name = "monitoring"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_integration_type.test", "id", "4"),
					resource.TestCheckResourceAttr("mcs_integration_type.test", "name", "monitoring"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_integration_type" "test" {
  name = "observability"
}`,
				Check: resource.TestCheckResourceAttr("mcs_integration_type.test", "name", "observability"),
			},
			{
				ResourceName:      "mcs_integration_type.test",
				ImportState:       true,
				ImportStateId:     "4",
				ImportStateVerify: true,
			},
			{
				ResourceName:  "mcs_integration_type.test",
				ImportState:   true,
				ImportStateId: "not-a-number",
				ExpectError:   regexpMustCompile(`Expected the numeric ID`),
			},
		},
	})
}

func TestAccIntegrationResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.OnStateful("/api/tenant/integrations", map[string]interface{}{
		"id": 7, "type": map[string]interface{}{"name": "PRTG"}, "subtype": "prtg",
		"username": "", "secret_expire": nil, "icon": "prtg.svg", "has_secret": "True", "has_key": false,
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_integration" "test" {
  name        = "prtg-prod"
  catalog_key = "prtg"
  baseurl     = "https://prtg.example.com"
  username    = "svc-mcs"
  secret      = "s3cret"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_integration.test", "id", "7"),
					resource.TestCheckResourceAttr("mcs_integration.test", "name", "prtg-prod"),
					resource.TestCheckResourceAttr("mcs_integration.test", "catalog_key", "prtg"),
					resource.TestCheckResourceAttr("mcs_integration.test", "type_name", "PRTG"),
					resource.TestCheckResourceAttr("mcs_integration.test", "subtype", "prtg"),
					resource.TestCheckResourceAttr("mcs_integration.test", "baseurl", "https://prtg.example.com"),
					resource.TestCheckResourceAttr("mcs_integration.test", "username", "svc-mcs"),
					resource.TestCheckResourceAttr("mcs_integration.test", "secret", "s3cret"),
					resource.TestCheckNoResourceAttr("mcs_integration.test", "key"),
					resource.TestCheckNoResourceAttr("mcs_integration.test", "secret_expire"),
					resource.TestCheckResourceAttr("mcs_integration.test", "icon", "prtg.svg"),
					resource.TestCheckResourceAttr("mcs_integration.test", "has_secret", "true"),
					resource.TestCheckResourceAttr("mcs_integration.test", "has_key", "false"),
					func(*terraform.State) error {
						for _, c := range mock.Calls() {
							if c.Method == http.MethodPost && strings.Contains(c.Body, `"secret":"s3cret"`) &&
								strings.Contains(c.Body, `"catalog_key":"prtg"`) {
								return nil
							}
						}
						return fmt.Errorf("create request did not include secret and catalog_key")
					},
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_integration" "test" {
  name          = "prtg-prod"
  catalog_key   = "prtg"
  type_id       = 2
  baseurl       = "https://prtg2.example.com"
  username      = "svc-mcs"
  secret        = "s3cret"
  key           = "api-key"
  secret_expire = "2027-01-01T00:00:00Z"
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("mcs_integration.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_integration.test", "type_id", "2"),
					resource.TestCheckResourceAttr("mcs_integration.test", "baseurl", "https://prtg2.example.com"),
					resource.TestCheckResourceAttr("mcs_integration.test", "key", "api-key"),
					resource.TestCheckResourceAttr("mcs_integration.test", "secret_expire", "2027-01-01T00:00:00Z"),
				),
			},
		},
	})
}

func TestAccLoadBalancerPoolDataSource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.OnList("/api/networking/loadbalancerpools", []map[string]interface{}{
		{"id": "lbp-001", "name": "web-pool", "description": "Web", "ip_pool": "ipp-001",
			"ip_pool_name": "lbpools-1", "prefix": "203.0.113.16/28", "status": "in_use", "customer": "acme",
			"assigned_object_type": "lbvserver", "assigned_object_id": "lbv-001", "assigned_object_label": "web-vip"},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_loadbalancer_pool" "web" {
  name = "web-pool"
}

data "mcs_loadbalancer_pool" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_loadbalancer_pool.web", "id", "lbp-001"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancer_pool.web", "description", "Web"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancer_pool.web", "ip_pool", "ipp-001"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancer_pool.web", "ip_pool_name", "lbpools-1"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancer_pool.web", "prefix", "203.0.113.16/28"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancer_pool.web", "status", "in_use"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancer_pool.web", "customer", "acme"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancer_pool.web", "assigned_object_type", "lbvserver"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancer_pool.web", "assigned_object_id", "lbv-001"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancer_pool.web", "assigned_object_label", "web-vip"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancer_pool.all", "pools.#", "1"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancer_pool.all", "pools.0.prefix", "203.0.113.16/28"),
				),
			},
		},
	})
}

func TestAccDomainQuotaDataSource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.OnList("/api/tenant/domainquotas", []map[string]interface{}{
		{"id": "dq-001", "name": "standard", "description": "Standard tier", "max_rules": 500, "max_objects": 1000,
			"max_objectgroups": 100, "max_services": 200, "max_servicegroups": 50, "max_networks": 10},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_domain_quota" "std" {
  name = "standard"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_domain_quota.std", "id", "dq-001"),
					resource.TestCheckResourceAttr("data.mcs_domain_quota.std", "description", "Standard tier"),
					resource.TestCheckResourceAttr("data.mcs_domain_quota.std", "max_rules", "500"),
					resource.TestCheckResourceAttr("data.mcs_domain_quota.std", "max_objects", "1000"),
					resource.TestCheckResourceAttr("data.mcs_domain_quota.std", "max_objectgroups", "100"),
					resource.TestCheckResourceAttr("data.mcs_domain_quota.std", "max_services", "200"),
					resource.TestCheckResourceAttr("data.mcs_domain_quota.std", "max_servicegroups", "50"),
					resource.TestCheckResourceAttr("data.mcs_domain_quota.std", "max_networks", "10"),
				),
			},
		},
	})
}

func TestAccLoadbalancingQuotaDataSource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.OnList("/api/tenant/loadbalancingquotas", []map[string]interface{}{
		{"id": "lq-001", "name": "standard", "description": "Standard tier", "max_policies_in_csvserver": 25},
		{"id": "lq-002", "name": "large", "description": "", "max_policies_in_csvserver": 100},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_loadbalancing_quota" "by_id" {
  id = "lq-002"
}

data "mcs_loadbalancing_quota" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_loadbalancing_quota.by_id", "name", "large"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancing_quota.by_id", "max_policies_in_csvserver", "100"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancing_quota.all", "loadbalancing_quotas.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_loadbalancing_quota.all", "loadbalancing_quotas.0.description", "Standard tier"),
				),
			},
		},
	})
}

func TestAccQuotaDataSource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.OnList("/api/tenant/quotas", []map[string]interface{}{
		{"id": "q-001", "name": "standard", "description": "Standard tier", "domain_quota": "dq-001", "loadbalancing_quota": "lq-001"},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_quota" "std" {
  name = "standard"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_quota.std", "id", "q-001"),
					resource.TestCheckResourceAttr("data.mcs_quota.std", "description", "Standard tier"),
					resource.TestCheckResourceAttr("data.mcs_quota.std", "domain_quota", "dq-001"),
					resource.TestCheckResourceAttr("data.mcs_quota.std", "loadbalancing_quota", "lq-001"),
				),
			},
		},
	})
}

func TestAccIntegrationTypeDataSource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.OnList("/api/tenant/integrationtypes", []map[string]interface{}{
		{"id": 1, "name": "monitoring"},
		{"id": 2, "name": "dns"},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_integration_type" "dns" {
  name = "dns"
}

data "mcs_integration_type" "by_id" {
  id = 1
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_integration_type.dns", "id", "2"),
					resource.TestCheckResourceAttr("data.mcs_integration_type.by_id", "name", "monitoring"),
				),
			},
		},
	})
}

func TestAccIntegrationDataSource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.OnList("/api/tenant/integrations", []map[string]interface{}{
		{"id": 7, "name": "prtg-prod", "type": map[string]interface{}{"name": "PRTG"}, "subtype": "prtg",
			"baseurl": "https://prtg.example.com", "username": "svc-mcs", "secret_expire": "2027-01-01T00:00:00Z",
			"icon": "prtg.svg", "has_secret": true, "has_key": "False"},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_integration" "prtg" {
  name = "prtg-prod"
}

data "mcs_integration" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_integration.prtg", "id", "7"),
					resource.TestCheckResourceAttr("data.mcs_integration.prtg", "type_name", "PRTG"),
					resource.TestCheckResourceAttr("data.mcs_integration.prtg", "subtype", "prtg"),
					resource.TestCheckResourceAttr("data.mcs_integration.prtg", "baseurl", "https://prtg.example.com"),
					resource.TestCheckResourceAttr("data.mcs_integration.prtg", "username", "svc-mcs"),
					resource.TestCheckResourceAttr("data.mcs_integration.prtg", "secret_expire", "2027-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("data.mcs_integration.prtg", "icon", "prtg.svg"),
					resource.TestCheckResourceAttr("data.mcs_integration.prtg", "has_secret", "true"),
					resource.TestCheckResourceAttr("data.mcs_integration.prtg", "has_key", "false"),
					resource.TestCheckResourceAttr("data.mcs_integration.all", "integrations.#", "1"),
				),
			},
		},
	})
}
