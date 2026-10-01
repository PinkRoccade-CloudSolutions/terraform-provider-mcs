package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func ingressClusterFixture(id, name, sla string, bandwidth int) map[string]interface{} {
	return map[string]interface{}{
		"id":                   id,
		"name":                 name,
		"slug":                 name,
		"sla":                  sla,
		"bandwidth":            bandwidth,
		"state":                "synced",
		"customer":             "cust-001",
		"ipaddress":            "pip-001",
		"firewall":             "xdp-001",
		"reverse_proxy":        7,
		"reverse_proxy_detail": map[string]interface{}{"id": 7, "name": "rp-ams-01"},
		"ipaddress_detail":     map[string]interface{}{"id": "pip-001", "address": "203.0.113.10", "type": "secureingress"},
		"created_at_timestamp": "2026-01-01T10:00:00Z",
		"updated_at_timestamp": "2026-01-02T10:00:00Z",
		"created_by_user":      42,
		"updated_by_user":      nil,
	}
}

// ---------------------------------------------------------------------------
// mcs_ingress_cluster data source
// ---------------------------------------------------------------------------

func TestAccIngressClusterDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var queries []string
	twoPageListMock(mock, "/api/secureingress/ingresscluster/",
		[]map[string]interface{}{ingressClusterFixture("ic-000", "webshop-staging", "bronze", 50)},
		[]map[string]interface{}{ingressClusterFixture("ic-001", "webshop", "gold", 1000)},
		&queries)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_ingress_cluster" "test" {
  name = "webshop"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "id", "ic-001"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "name", "webshop"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "sla", "gold"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "bandwidth", "1000"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "customer", "cust-001"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "ipaddress", "pip-001"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "firewall", "xdp-001"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "slug", "webshop"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "state", "synced"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "reverse_proxy", "7"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "reverse_proxy_name", "rp-ams-01"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "ipaddress_address", "203.0.113.10"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "ipaddress_type", "secureingress"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "created_at_timestamp", "2026-01-01T10:00:00Z"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "updated_at_timestamp", "2026-01-02T10:00:00Z"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "created_by_user", "42"),
					resource.TestCheckNoResourceAttr("data.mcs_ingress_cluster.test", "updated_by_user"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.test", "ingress_clusters.#", "0"),
					checkQueryContains(&queries, "name__icontains=webshop"),
				),
			},
		},
	})
}

func TestAccIngressClusterDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var queries []string
	twoPageListMock(mock, "/api/secureingress/ingresscluster/",
		[]map[string]interface{}{ingressClusterFixture("ic-001", "webshop", "gold", 1000)},
		[]map[string]interface{}{ingressClusterFixture("ic-002", "portal", "gold", 1000)},
		&queries)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_ingress_cluster" "gold" {
  sla       = "gold"
  bandwidth = 1000
  ipaddress = "pip-001"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.gold", "ingress_clusters.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.gold", "ingress_clusters.0.id", "ic-001"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.gold", "ingress_clusters.0.reverse_proxy_name", "rp-ams-01"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.gold", "ingress_clusters.1.name", "portal"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.gold", "ingress_clusters.1.ipaddress_address", "203.0.113.10"),
					resource.TestCheckResourceAttr("data.mcs_ingress_cluster.gold", "sla", "gold"),
					resource.TestCheckNoResourceAttr("data.mcs_ingress_cluster.gold", "id"),
					checkQueryContains(&queries, "sla=gold"),
					checkQueryContains(&queries, "bandwidth=1000"),
					checkQueryContains(&queries, "ipaddress=pip-001"),
				),
			},
		},
	})
}
