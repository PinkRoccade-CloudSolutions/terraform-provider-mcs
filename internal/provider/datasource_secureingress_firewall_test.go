package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func secureIngressFirewallFixture(id, name string, description interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":                   id,
		"name":                 name,
		"description":          description,
		"filter_json":          `{"allow": ["10.0.0.0/8"]}`,
		"customer":             "cust-001",
		"created_at_timestamp": "2026-01-01T10:00:00Z",
		"updated_at_timestamp": "2026-01-02T10:00:00Z",
		"created_by_user":      42,
		"updated_by_user":      nil,
	}
}

// checkQueryContains asserts that some recorded query string contains want.
func checkQueryContains(queries *[]string, want string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		for _, q := range *queries {
			if strings.Contains(q, want) {
				return nil
			}
		}
		return fmt.Errorf("no request with query containing %q (got %v)", want, *queries)
	}
}

// ---------------------------------------------------------------------------
// mcs_secureingress_firewall data source
// ---------------------------------------------------------------------------

func TestAccSecureIngressFirewallDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var queries []string
	// name__icontains returns partial matches; the data source must pick the exact one.
	twoPageListMock(mock, "/api/secureingress/firewall/",
		[]map[string]interface{}{secureIngressFirewallFixture("xdp-000", "edge-filter-old", nil)},
		[]map[string]interface{}{secureIngressFirewallFixture("xdp-001", "edge-filter", "Edge traffic filter")},
		&queries)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_secureingress_firewall" "test" {
  name = "edge-filter"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.test", "id", "xdp-001"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.test", "name", "edge-filter"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.test", "description", "Edge traffic filter"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.test", "filter_json", `{"allow": ["10.0.0.0/8"]}`),
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.test", "customer", "cust-001"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.test", "created_at_timestamp", "2026-01-01T10:00:00Z"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.test", "updated_at_timestamp", "2026-01-02T10:00:00Z"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.test", "created_by_user", "42"),
					resource.TestCheckNoResourceAttr("data.mcs_secureingress_firewall.test", "updated_by_user"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.test", "secureingress_firewalls.#", "0"),
					checkQueryContains(&queries, "name__icontains=edge-filter"),
				),
			},
		},
	})
}

func TestAccSecureIngressFirewallDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	twoPageListMock(mock, "/api/secureingress/firewall/",
		[]map[string]interface{}{secureIngressFirewallFixture("xdp-001", "edge-filter", "Edge traffic filter")},
		[]map[string]interface{}{secureIngressFirewallFixture("xdp-002", "api-filter", nil)},
		nil)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_secureingress_firewall" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.all", "secureingress_firewalls.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.all", "secureingress_firewalls.0.id", "xdp-001"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.all", "secureingress_firewalls.0.description", "Edge traffic filter"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.all", "secureingress_firewalls.1.name", "api-filter"),
					resource.TestCheckNoResourceAttr("data.mcs_secureingress_firewall.all", "secureingress_firewalls.1.description"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_firewall.all", "secureingress_firewalls.1.customer", "cust-001"),
				),
			},
		},
	})
}
