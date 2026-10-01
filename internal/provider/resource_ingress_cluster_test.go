package provider

import (
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ---------------------------------------------------------------------------
// mcs_ingress_cluster
// ---------------------------------------------------------------------------

func TestAccIngressClusterResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	addresses := map[string]string{"pip-001": "203.0.113.10", "pip-002": "203.0.113.20"}
	writes := 0
	store := newSingleObjectStore(mock, "/api/secureingress/ingresscluster/", "ic-001",
		map[string]interface{}{
			"sla": "bronze", "bandwidth": 50,
			"created_at_timestamp": "2026-01-01T10:00:00Z", "created_by_user": 42,
		},
		func(obj, _ map[string]interface{}) {
			writes++
			obj["slug"] = strings.ToLower(obj["name"].(string))
			obj["state"] = "synced"
			obj["reverse_proxy"] = 7
			obj["reverse_proxy_detail"] = map[string]interface{}{"id": 7, "name": "rp-ams-01"}
			ip := obj["ipaddress"].(string)
			obj["ipaddress_detail"] = map[string]interface{}{"id": ip, "address": addresses[ip], "type": "secureingress"}
			obj["updated_at_timestamp"] = []string{"", "2026-01-01T10:00:00Z", "2026-01-02T10:00:00Z", "2026-01-03T10:00:00Z"}[min(writes, 3)]
			obj["updated_by_user"] = 42 + writes
		})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				// sla and bandwidth unset: the API defaults are adopted.
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_ingress_cluster" "test" {
  name      = "Webshop"
  customer  = "cust-001"
  ipaddress = "pip-001"
  firewall  = "xdp-001"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "id", "ic-001"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "name", "Webshop"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "sla", "bronze"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "bandwidth", "50"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "customer", "cust-001"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "ipaddress", "pip-001"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "firewall", "xdp-001"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "slug", "webshop"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "state", "synced"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "reverse_proxy", "7"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "reverse_proxy_name", "rp-ams-01"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "ipaddress_address", "203.0.113.10"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "ipaddress_type", "secureingress"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "created_at_timestamp", "2026-01-01T10:00:00Z"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "updated_at_timestamp", "2026-01-01T10:00:00Z"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "created_by_user", "42"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "updated_by_user", "43"),
					store.checkWrite(http.MethodPost, "/api/secureingress/ingresscluster/"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_ingress_cluster" "test" {
  name      = "Webshop-Prod"
  sla       = "gold"
  bandwidth = 1000
  customer  = "cust-001"
  ipaddress = "pip-002"
  firewall  = "xdp-002"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "id", "ic-001"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "name", "Webshop-Prod"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "sla", "gold"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "bandwidth", "1000"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "ipaddress", "pip-002"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "firewall", "xdp-002"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "slug", "webshop-prod"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "ipaddress_address", "203.0.113.20"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "created_at_timestamp", "2026-01-01T10:00:00Z"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "updated_at_timestamp", "2026-01-02T10:00:00Z"),
					resource.TestCheckResourceAttr("mcs_ingress_cluster.test", "updated_by_user", "44"),
					store.checkWrite(http.MethodPut, "/api/secureingress/ingresscluster/ic-001/"),
					store.checkStored("bandwidth", 1000),
				),
			},
			{
				ResourceName:      "mcs_ingress_cluster.test",
				ImportState:       true,
				ImportStateId:     "ic-001",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIngressClusterResource_InvalidBandwidth(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories("http://127.0.0.1:1"),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock("http://127.0.0.1:1") + `
resource "mcs_ingress_cluster" "test" {
  name      = "Webshop"
  bandwidth = 200
  customer  = "cust-001"
  ipaddress = "pip-001"
  firewall  = "xdp-001"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
		},
	})
}
