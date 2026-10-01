package provider

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func netCheckLastBody(m *mockAPIServer, method, prefix string, wants ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		bodies := netCallBodies(m, method, prefix)
		if len(bodies) == 0 {
			return fmt.Errorf("no %s call to %s recorded", method, prefix)
		}
		last := bodies[len(bodies)-1]
		for _, w := range wants {
			if !strings.Contains(last, w) {
				return fmt.Errorf("%s body %s does not contain %s", method, last, w)
			}
		}
		return nil
	}
}

func netCheckCallCount(m *mockAPIServer, method, prefix string, want int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if got := len(netCallBodies(m, method, prefix)); got != want {
			return fmt.Errorf("expected %d %s calls to %s, got %d", want, method, prefix, got)
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// mcs_nat_translation
// ---------------------------------------------------------------------------

func TestAccNATTranslationResource_NullableAndSnat(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	netStatefulMock(mock, "/api/networking/nattranslations/", map[string]interface{}{
		"id": "nat-001", "private_ip": "192.168.1.1", "translation": "203.0.113.1:80 -> 192.168.1.1:8080",
		"state": "unsynced", "snat_source_addresses": nil, "snat_translated_addresses": nil,
	}, func(_ string, obj map[string]interface{}) {
		obj["state"] = "synced"
	})

	res := "mcs_nat_translation.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_nat_translation" "test" {
  public_ip        = "pip-uuid-1"
  interface        = "if-uuid-1"
  firewall         = "fw-uuid-1"
  translation_type = "port_forward"
  public_port      = 80
  private_port     = 8080
  protocol         = "tcp"
  customer         = "acme"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "id", "nat-001"),
					resource.TestCheckResourceAttr(res, "public_port", "80"),
					resource.TestCheckResourceAttr(res, "private_port", "8080"),
					resource.TestCheckResourceAttr(res, "translation", "203.0.113.1:80 -> 192.168.1.1:8080"),
					resource.TestCheckResourceAttr(res, "state", "synced"),
					resource.TestCheckResourceAttr(res, "private_ip", "192.168.1.1"),
					resource.TestCheckResourceAttr(res, "enabled", "true"),
					resource.TestCheckResourceAttr(res, "description", ""),
					resource.TestCheckNoResourceAttr(res, "snat_source_addresses.#"),
					netCheckLastBody(mock, "POST", "/api/networking/nattranslations", `"enabled":true`, `"public_port":80`),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_nat_translation" "test" {
  interface                 = "if-uuid-1"
  firewall                  = "fw-uuid-1"
  translation_type          = "snat"
  customer                  = "acme"
  protocol                  = ""
  snat_source_addresses     = ["10.0.0.0/24", "10.0.1.0/24"]
  snat_translated_addresses = ["203.0.113.5"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "id", "nat-001"),
					resource.TestCheckNoResourceAttr(res, "public_ip"),
					resource.TestCheckNoResourceAttr(res, "public_port"),
					resource.TestCheckNoResourceAttr(res, "private_port"),
					resource.TestCheckResourceAttr(res, "translation_type", "snat"),
					resource.TestCheckResourceAttr(res, "snat_source_addresses.#", "2"),
					resource.TestCheckResourceAttr(res, "snat_source_addresses.1", "10.0.1.0/24"),
					resource.TestCheckResourceAttr(res, "snat_translated_addresses.0", "203.0.113.5"),
					netCheckLastBody(mock, "PUT", "/api/networking/nattranslations/nat-001",
						`"public_ip":null`, `"public_port":null`, `"private_port":null`,
						`"snat_source_addresses":["10.0.0.0/24","10.0.1.0/24"]`),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_nat_translation" "test" {
  interface        = "if-uuid-1"
  firewall         = "fw-uuid-1"
  translation_type = "snat"
  customer         = "acme"
  protocol         = ""
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(res, "snat_source_addresses.#"),
					netCheckLastBody(mock, "PUT", "/api/networking/nattranslations/nat-001",
						`"snat_source_addresses":[]`, `"snat_translated_addresses":[]`),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_nat_translation" "test" {
  interface        = "if-uuid-1"
  firewall         = "fw-uuid-2"
  translation_type = "snat"
  customer         = "acme"
  protocol         = ""
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "firewall", "fw-uuid-2"),
					netCheckCallCount(mock, "POST", "/api/networking/nattranslations", 2),
					netCheckCallCount(mock, "DELETE", "/api/networking/nattranslations", 1),
				),
			},
		},
	})
}

func TestAccNATTranslationResource_Validators(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_nat_translation" "test" {
  firewall         = "fw-uuid-1"
  translation_type = "many_to_many"
  customer         = "acme"
}`,
				ExpectError: regexpMustCompile(`translation_type`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_nat_translation" "test" {
  firewall         = "fw-uuid-1"
  translation_type = "port_forward"
  public_port      = 4294967296
  customer         = "acme"
}`,
				ExpectError: regexpMustCompile(`public_port`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_nat_translation" "test" {
  firewall         = "fw-uuid-1"
  translation_type = "port_forward"
  protocol         = "icmp"
  customer         = "acme"
}`,
				ExpectError: regexpMustCompile(`protocol`),
			},
		},
	})
}

func TestAccNATTranslationDataSource_Paginated(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var queries []string
	netPagedMock(mock, "/api/networking/nattranslations/", [][]map[string]interface{}{
		{{"id": "nat-001", "public_ip": nil, "interface": "if-1", "firewall": "fw-1", "translation_type": "snat",
			"public_port": nil, "private_port": nil, "customer": "acme", "state": "synced", "enabled": true,
			"snat_source_addresses": []string{"10.0.0.0/24"}, "snat_translated_addresses": "203.0.113.5, 203.0.113.6"}},
		{{"id": "nat-002", "public_ip": "pip-2", "interface": nil, "firewall": "fw-1", "translation_type": "port_forward",
			"public_port": 443, "private_port": 8443, "protocol": "tcp", "customer": "acme", "state": "error", "enabled": false}},
	}, &queries)

	ds := "data.mcs_nat_translation.all"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_nat_translation" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ds, "nat_translations.#", "2"),
					resource.TestCheckNoResourceAttr(ds, "nat_translations.0.public_ip"),
					resource.TestCheckNoResourceAttr(ds, "nat_translations.0.public_port"),
					resource.TestCheckResourceAttr(ds, "nat_translations.0.snat_source_addresses.0", "10.0.0.0/24"),
					resource.TestCheckResourceAttr(ds, "nat_translations.0.snat_translated_addresses.#", "2"),
					resource.TestCheckResourceAttr(ds, "nat_translations.0.snat_translated_addresses.1", "203.0.113.6"),
					resource.TestCheckResourceAttr(ds, "nat_translations.1.public_ip", "pip-2"),
					resource.TestCheckNoResourceAttr(ds, "nat_translations.1.interface"),
					resource.TestCheckResourceAttr(ds, "nat_translations.1.public_port", "443"),
					resource.TestCheckResourceAttr(ds, "nat_translations.1.state", "error"),
					resource.TestCheckResourceAttr(ds, "nat_translations.1.snat_source_addresses.#", "0"),
				),
			},
		},
	})
	for _, q := range queries {
		if strings.Contains(q, "page_size") {
			t.Errorf("unexpected page_size query: %q", q)
		}
	}
}

// ---------------------------------------------------------------------------
// mcs_public_ip_address
// ---------------------------------------------------------------------------

func TestAccPublicIPAddressResource_Assigned(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	netStatefulMock(mock, "/api/networking/publicipaddresss/", map[string]interface{}{
		"id": "pip-001", "status": "available", "customer": "acme",
	}, func(_ string, obj map[string]interface{}) {
		if obj["ip_address"] == nil {
			obj["ip_address"] = "203.0.113.10"
		}
		if obj["pool"] == nil {
			obj["pool"] = "pool-auto"
		}
		obj["status"] = "assigned"
	})

	res := "mcs_public_ip_address.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_public_ip_address" "test" {
  type        = "secureingress"
  description = "Ingress IP"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "ip_address", "203.0.113.10"),
					resource.TestCheckResourceAttr(res, "pool", "pool-auto"),
					resource.TestCheckResourceAttr(res, "status", "assigned"),
					resource.TestCheckResourceAttr(res, "customer", "acme"),
					resource.TestCheckResourceAttr(res, "type", "secureingress"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_public_ip_address" "test" {
  type = "secureingress"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "description", ""),
					resource.TestCheckResourceAttr(res, "ip_address", "203.0.113.10"),
					netCheckLastBody(mock, "PUT", "/api/networking/publicipaddresss/pip-001",
						`"description":""`, `"ip_address":"203.0.113.10"`, `"pool":"pool-auto"`),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_public_ip_address" "test" {
  type       = "secureingress"
  pool       = "pool-other"
  ip_address = "198.51.100.7"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "pool", "pool-other"),
					resource.TestCheckResourceAttr(res, "ip_address", "198.51.100.7"),
					netCheckCallCount(mock, "DELETE", "/api/networking/publicipaddresss", 1),
					netCheckLastBody(mock, "POST", "/api/networking/publicipaddresss", `"ip_address":"198.51.100.7"`),
				),
			},
		},
	})
}

func TestAccPublicIPAddressResource_TypeValidator(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_public_ip_address" "test" {
  type = "public"
}`,
				ExpectError: regexpMustCompile(`type`),
			},
		},
	})
}

func TestAccPublicIPAddressDataSource_Paginated(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var queries []string
	netPagedMock(mock, "/api/networking/publicipaddresss/", [][]map[string]interface{}{
		{{"id": "pip-001", "ip_address": nil, "pool": "pool-1", "status": "reserved", "type": "nat"}},
		{{"id": "pip-002", "ip_address": "203.0.113.20", "pool": nil, "status": "assigned", "type": "secureingress"}},
	}, &queries)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_public_ip_address" "test" {
  ip_address = "203.0.113.20"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "id", "pip-002"),
					resource.TestCheckResourceAttr("data.mcs_public_ip_address.test", "type", "secureingress"),
					resource.TestCheckNoResourceAttr("data.mcs_public_ip_address.test", "pool"),
				),
			},
		},
	})
	if len(queries) == 0 || queries[0] != "ip_address__icontains=203.0.113.20" {
		t.Errorf("unexpected first query: %v", queries)
	}
}

// ---------------------------------------------------------------------------
// mcs_network data source
// ---------------------------------------------------------------------------

func TestAccNetworkDataSource_DomainAndPagination(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var queries []string
	netPagedMock(mock, "/api/v3/networking/networks/", [][]map[string]interface{}{
		{{"id": "net-001", "name": "mgmt", "vlanid": 100, "domain": 5,
			"domain_detail": map[string]interface{}{"id": 5, "vdom_id": "vd5", "name": "prod"}}},
		{{"id": "net-002", "name": "data", "vlanid": 200, "domain": nil, "domain_detail": nil,
			"description": "Data net", "ipv4_prefix": "10.1.0.0/24", "ipv4_address": "10.1.0.1"}},
	}, &queries)

	ds := "data.mcs_network.all"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_network" "all" {
  domain = 5
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ds, "networks.#", "2"),
					resource.TestCheckResourceAttr(ds, "domain", "5"),
					resource.TestCheckResourceAttr(ds, "networks.0.vlan_id", "100"),
					resource.TestCheckResourceAttr(ds, "networks.0.domain", "5"),
					resource.TestCheckResourceAttr(ds, "networks.0.domain_name", "prod"),
					resource.TestCheckResourceAttr(ds, "networks.0.domain_vdom_id", "vd5"),
					resource.TestCheckNoResourceAttr(ds, "networks.1.domain"),
					resource.TestCheckNoResourceAttr(ds, "networks.1.domain_name"),
					resource.TestCheckResourceAttr(ds, "networks.1.description", "Data net"),
					resource.TestCheckResourceAttr(ds, "networks.1.ipv4_address", "10.1.0.1"),
				),
			},
		},
	})
	if len(queries) == 0 || queries[0] != "domain=5" {
		t.Errorf("unexpected first query: %v", queries)
	}
}

// ---------------------------------------------------------------------------
// mcs_zone data source
// ---------------------------------------------------------------------------

func TestAccZoneDataSource_FieldsAndPagination(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var queries []string
	netPagedMock(mock, "/api/networking/zones/", [][]map[string]interface{}{
		{{"id": 1, "uuid": "zone-uuid-1", "name": "dmz-old", "loadbalancers": []interface{}{}, "edge_root_networks": []string{}}},
		{{"id": 2, "uuid": "zone-uuid-2", "name": "dmz",
			"loadbalancers":      []map[string]interface{}{{"id": "lb-1", "alias": "lb-dmz"}},
			"edge_root_networks": []string{"net-a", "net-b"}}},
	}, &queries)

	ds := "data.mcs_zone.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_zone" "test" {
  name = "dmz"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ds, "id", "2"),
					resource.TestCheckResourceAttr(ds, "uuid", "zone-uuid-2"),
					resource.TestCheckResourceAttr(ds, "loadbalancers.0.name", "lb-dmz"),
					resource.TestCheckResourceAttr(ds, "edge_root_networks.#", "2"),
					resource.TestCheckResourceAttr(ds, "edge_root_networks.1", "net-b"),
				),
			},
		},
	})
	if len(queries) == 0 || queries[0] != "name__icontains=dmz" {
		t.Errorf("unexpected first query: %v", queries)
	}
}

// ---------------------------------------------------------------------------
// mcs_site_to_site_vpn
// ---------------------------------------------------------------------------

func TestAccSiteToSiteVPNResource_FieldsAndUpdate(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	updates := 0
	netStatefulMock(mock, "/api/vpn/site_to_site/", map[string]interface{}{
		"id": 9, "uuid": "vpn-uuid-009", "name": "", "state": "", "last_status": "", "resets": 0,
		"last_check": nil, "last_reset": nil, "customer": "acme", "firewall": nil,
		"created_at_timestamp": "2025-01-01T00:00:00Z", "updated_at_timestamp": "2025-01-01T00:00:00Z",
		"created_by_user": 3, "updated_by_user": 3,
	}, func(method string, obj map[string]interface{}) {
		if method == "PUT" {
			updates++
			obj["updated_at_timestamp"] = fmt.Sprintf("2025-01-0%dT00:00:00Z", updates+1)
			obj["updated_by_user"] = 4
		}
	})

	res := "mcs_site_to_site_vpn.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_site_to_site_vpn" "test" {
  domain   = 11
  tenant   = 22
  firewall = "fw-uuid-1"
  contact  = [1, 2]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "id", "9"),
					resource.TestCheckResourceAttr(res, "name", ""),
					resource.TestCheckResourceAttr(res, "domain", "11"),
					resource.TestCheckResourceAttr(res, "tenant", "22"),
					resource.TestCheckResourceAttr(res, "customer", "acme"),
					resource.TestCheckResourceAttr(res, "firewall", "fw-uuid-1"),
					resource.TestCheckResourceAttr(res, "contact.#", "2"),
					resource.TestCheckResourceAttr(res, "contact.1", "2"),
					resource.TestCheckNoResourceAttr(res, "last_check"),
					resource.TestCheckResourceAttr(res, "created_by_user", "3"),
					resource.TestCheckResourceAttr(res, "created_at_timestamp", "2025-01-01T00:00:00Z"),
					netCheckLastBody(mock, "POST", "/api/vpn/site_to_site", `"domain":11`, `"tenant":22`, `"name":""`),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_site_to_site_vpn" "test" {
  name     = "hq"
  domain   = 11
  tenant   = 22
  firewall = "fw-uuid-1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "name", "hq"),
					resource.TestCheckNoResourceAttr(res, "contact.#"),
					resource.TestCheckResourceAttr(res, "updated_at_timestamp", "2025-01-02T00:00:00Z"),
					resource.TestCheckResourceAttr(res, "updated_by_user", "4"),
					netCheckLastBody(mock, "PUT", "/api/vpn/site_to_site/9", `"contact":[]`, `"name":"hq"`),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_site_to_site_vpn" "test" {
  name     = "hq"
  domain   = 12
  tenant   = 22
  firewall = "fw-uuid-1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "domain", "12"),
					netCheckCallCount(mock, "DELETE", "/api/vpn/site_to_site", 1),
				),
			},
		},
	})
}

func TestAccSiteToSiteVPNDataSource_ExactNameFilterAndPagination(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var queries []string
	netPagedMock(mock, "/api/vpn/site_to_site/", [][]map[string]interface{}{
		{{"id": 1, "uuid": "u1", "name": "other", "domain": 1, "tenant": 1}},
		{{"id": 2, "uuid": "u2", "name": "hq", "domain": 11, "tenant": 22, "customer": "acme", "firewall": nil,
			"contact": []int{5}, "last_check": nil, "last_reset": "2025-02-01T00:00:00Z", "created_by_user": 3}},
	}, &queries)

	ds := "data.mcs_site_to_site_vpn.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_site_to_site_vpn" "test" {
  name = "hq"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ds, "id", "2"),
					resource.TestCheckResourceAttr(ds, "domain", "11"),
					resource.TestCheckResourceAttr(ds, "tenant", "22"),
					resource.TestCheckResourceAttr(ds, "customer", "acme"),
					resource.TestCheckNoResourceAttr(ds, "firewall"),
					resource.TestCheckResourceAttr(ds, "contact.0", "5"),
					resource.TestCheckNoResourceAttr(ds, "last_check"),
					resource.TestCheckResourceAttr(ds, "last_reset", "2025-02-01T00:00:00Z"),
					resource.TestCheckResourceAttr(ds, "created_by_user", "3"),
				),
			},
		},
	})
	if len(queries) == 0 || queries[0] != "name=hq" {
		t.Errorf("unexpected first query: %v", queries)
	}
}

// ---------------------------------------------------------------------------
// mcs_monitor_ip
// ---------------------------------------------------------------------------

func TestAccMonitorIPResource_OptionalStrings(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	netStatefulMock(mock, "/api/dbl/monitorip/", map[string]interface{}{
		"id": "mip-002", "timestamp": "2025-01-01T00:00:00Z", "last_check_timestamp": "2025-01-01",
		"notify_email": "", "comment": "",
	}, nil)

	res := "mcs_monitor_ip.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_monitor_ip" "test" {
  ipaddress    = "10.0.0.100"
  customer     = "acme"
  comment      = "watch"
  notify_email = "ops@example.com"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "comment", "watch"),
					resource.TestCheckResourceAttr(res, "notify_email", "ops@example.com"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_monitor_ip" "test" {
  ipaddress = "10.0.0.100"
  customer  = "acme"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "comment", ""),
					resource.TestCheckResourceAttr(res, "notify_email", ""),
					netCheckLastBody(mock, "PUT", "/api/dbl/monitorip/mip-002", `"comment":""`, `"notify_email":""`),
				),
			},
		},
	})
}

func TestAccMonitorIPDataSource_Paginated(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var queries []string
	netPagedMock(mock, "/api/dbl/monitorip/", [][]map[string]interface{}{
		{{"id": "mip-1", "ipaddress": "10.0.0.1", "customer": "acme"}},
		{{"id": "mip-2", "ipaddress": "10.0.0.2", "customer": "acme", "comment": "c"}},
	}, &queries)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_monitor_ip" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_monitor_ip.all", "monitor_ips.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_monitor_ip.all", "monitor_ips.1.comment", "c"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_ippool / mcs_networkpool data sources
// ---------------------------------------------------------------------------

func TestAccIPPoolDataSource_FiltersAndPagination(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var queries []string
	netPagedMock(mock, "/api/networking/ippools/", [][]map[string]interface{}{
		{{"id": "pool-1", "name": "A", "subnet": "203.0.113.0/28", "type": "nat", "customer": "acme", "total_ips": "16", "free_ips": "10"}},
		{{"id": "pool-2", "name": "B", "subnet": "198.51.100.0/24", "type": "nat", "customer": "acme", "total_ips": 256, "free_ips": 250}},
	}, &queries)

	ds := "data.mcs_ippool.all"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_ippool" "all" {
  customer = "acme"
  type     = "nat"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ds, "ip_pools.#", "2"),
					resource.TestCheckResourceAttr(ds, "customer", "acme"),
					resource.TestCheckResourceAttr(ds, "ip_pools.0.type", "nat"),
					resource.TestCheckResourceAttr(ds, "ip_pools.0.total_ips", "16"),
					resource.TestCheckResourceAttr(ds, "ip_pools.0.free_ips", "10"),
					resource.TestCheckResourceAttr(ds, "ip_pools.1.total_ips", "256"),
					resource.TestCheckResourceAttr(ds, "ip_pools.1.free_ips", "250"),
				),
			},
		},
	})
	if len(queries) == 0 || queries[0] != "customer=acme&type=nat" {
		t.Errorf("unexpected first query: %v", queries)
	}
}

func TestAccNetworkPoolDataSource_Paginated(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var queries []string
	netPagedMock(mock, "/api/networking/networkpools/", [][]map[string]interface{}{
		{{"id": "np-1", "name": "LAN Pool", "network": "10.0.0.0/8", "type": "lan", "enabled": true}},
		{{"id": "np-2", "name": "Transit Pool", "network": "100.64.0.0/10", "type": "transit", "enabled": false}},
	}, &queries)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_networkpool" "test" {
  name = "Transit Pool"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_networkpool.test", "id", "np-2"),
					resource.TestCheckResourceAttr("data.mcs_networkpool.test", "type", "transit"),
				),
			},
		},
	})
}
