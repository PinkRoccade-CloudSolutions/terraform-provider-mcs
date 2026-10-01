package provider

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ---------------------------------------------------------------------------
// mcs_secureingress_firewall
// ---------------------------------------------------------------------------

func TestAccSecureIngressFirewallResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	writes := 0
	store := newSingleObjectStore(mock, "/api/secureingress/firewall/", "xdp-001",
		map[string]interface{}{"description": nil, "created_at_timestamp": "2026-01-01T10:00:00Z", "created_by_user": 42},
		func(obj, req map[string]interface{}) {
			writes++
			obj["updated_at_timestamp"] = []string{"", "2026-01-01T10:00:00Z", "2026-01-02T10:00:00Z", "2026-01-03T10:00:00Z"}[min(writes, 3)]
			obj["updated_by_user"] = 42 + writes
			// The API reformats the JSON string; the provider must not report a diff for that.
			if s, ok := req["filter_json"].(string); ok {
				var buf bytes.Buffer
				_ = json.Indent(&buf, []byte(s), "", "    ")
				obj["filter_json"] = buf.String()
			}
		})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_secureingress_firewall" "test" {
  name        = "edge-filter"
  customer    = "cust-001"
  filter_json = jsonencode({ allow = ["10.0.0.0/8"], block_countries = ["XX"] })
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "id", "xdp-001"),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "name", "edge-filter"),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "customer", "cust-001"),
					resource.TestCheckNoResourceAttr("mcs_secureingress_firewall.test", "description"),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "filter_json", `{"allow":["10.0.0.0/8"],"block_countries":["XX"]}`),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "created_at_timestamp", "2026-01-01T10:00:00Z"),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "updated_at_timestamp", "2026-01-01T10:00:00Z"),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "created_by_user", "42"),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "updated_by_user", "43"),
					// filter_json is sent as a JSON-encoded string, as the spec defines it.
					store.checkStored("filter_json", "{\n    \"allow\": [\n        \"10.0.0.0/8\"\n    ],\n    \"block_countries\": [\n        \"XX\"\n    ]\n}"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_secureingress_firewall" "test" {
  name        = "edge-filter-v2"
  description = "Blocks unwanted traffic"
  customer    = "cust-001"
  filter_json = jsonencode({ allow = ["10.0.0.0/8", "192.168.0.0/16"] })
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "id", "xdp-001"),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "name", "edge-filter-v2"),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "description", "Blocks unwanted traffic"),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "filter_json", `{"allow":["10.0.0.0/8","192.168.0.0/16"]}`),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "created_at_timestamp", "2026-01-01T10:00:00Z"),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "updated_at_timestamp", "2026-01-02T10:00:00Z"),
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "updated_by_user", "44"),
					store.checkWrite(http.MethodPut, "/api/secureingress/firewall/xdp-001/"),
				),
			},
			{
				ResourceName:      "mcs_secureingress_firewall.test",
				ImportState:       true,
				ImportStateId:     "xdp-001",
				ImportStateVerify: true,
				// The API's formatting of filter_json differs from jsonencode(); both are semantically equal.
				ImportStateVerifyIgnore: []string{"filter_json"},
			},
		},
	})
}

func TestAccSecureIngressFirewallResource_CustomerRequiresReplace(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	store := newSingleObjectStore(mock, "/api/secureingress/firewall/", "xdp-001",
		map[string]interface{}{"description": nil, "created_at_timestamp": "t0", "updated_at_timestamp": "t0", "created_by_user": nil, "updated_by_user": nil}, nil)

	cfg := func(customer string) string {
		return providerConfigBlock(mock.URL()) + `
resource "mcs_secureingress_firewall" "test" {
  name        = "edge-filter"
  customer    = "` + customer + `"
  filter_json = jsonencode({})
}`
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{Config: cfg("cust-001")},
			{
				Config: cfg("cust-002"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_secureingress_firewall.test", "customer", "cust-002"),
					resource.TestCheckNoResourceAttr("mcs_secureingress_firewall.test", "created_by_user"),
					store.checkWrite(http.MethodDelete, "/api/secureingress/firewall/xdp-001/"),
				),
			},
		},
	})
}
