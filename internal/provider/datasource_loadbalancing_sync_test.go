package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// twoPageMock serves items over two pages (page 2 via an absolute `next` link) and records
// every query string it receives.
func twoPageMock(mock *mockAPIServer, path string, page1, page2 []map[string]interface{}) func() []string {
	var mu sync.Mutex
	var queries []string
	mock.On(path, func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		mu.Lock()
		queries = append(queries, r.URL.RawQuery)
		mu.Unlock()
		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"count": len(page1) + len(page2), "next": nil, "results": page2,
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count": len(page1) + len(page2), "next": mock.URL() + path + "?page=2", "results": page1,
		})
	})
	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), queries...)
	}
}

func TestLbSync_DataSources_Pagination(t *testing.T) {
	cases := []struct {
		typeName string
		path     string
		listAttr string
		byName   bool
		page1    map[string]interface{}
		page2    map[string]interface{}
		checks   map[string]string
	}{
		{
			typeName: "mcs_certificate", path: "/api/loadbalancing/certificate/", listAttr: "certificates", byName: true,
			page1:  map[string]interface{}{"id": "c-1", "name": "first", "loadbalancer": "lb-1", "protected": false},
			page2:  map[string]interface{}{"id": "c-2", "name": "target", "loadbalancer": "lb-1", "customer": "cust-1", "protected": true},
			checks: map[string]string{"customer": "cust-1", "protected": "true"},
		},
		{
			typeName: "mcs_csv_server", path: "/api/loadbalancing/csvserver/", listAttr: "csv_servers", byName: true,
			page1: map[string]interface{}{"id": "csv-1", "name": "first", "ufname": "f", "type": "http"},
			page2: map[string]interface{}{
				"id": "csv-2", "name": "target", "ufname": "t", "type": "ssl", "port": 443,
				"ca_certificate": []string{"ca-1"}, "clientauth": true, "clientcert": "Mandatory",
			},
			checks: map[string]string{"ca_certificate.#": "1", "ca_certificate.0": "ca-1", "clientauth": "true", "clientcert": "Mandatory", "port": "443"},
		},
		{
			typeName: "mcs_lbv_server", path: "/api/loadbalancing/lbvserver/", listAttr: "lbv_servers", byName: true,
			page1:  map[string]interface{}{"id": "lbv-1", "name": "first", "servicegroup": []string{}},
			page2:  map[string]interface{}{"id": "lbv-2", "name": "target", "type": "ssl", "servicegroup": []string{"sg-1"}, "ca_certificate": []string{"ca-9"}},
			checks: map[string]string{"ca_certificate.0": "ca-9", "type": "ssl", "servicegroup.0": "sg-1"},
		},
		{
			typeName: "mcs_lb_servicegroup", path: "/api/loadbalancing/lbservicegroup/", listAttr: "lb_servicegroups", byName: true,
			page1: map[string]interface{}{"id": "sg-1", "name": "first", "type": "HTTP"},
			page2: map[string]interface{}{
				"id": "sg-2", "name": "target", "type": "SSL", "monitors": []string{"mon-1", "mon-2"},
				"client_certificate": "cert-1", "cip": "ENABLED", "cipheader": "X-Forwarded-For",
			},
			checks: map[string]string{"monitors.#": "2", "monitors.1": "mon-2", "client_certificate": "cert-1", "cip": "ENABLED", "cipheader": "X-Forwarded-For"},
		},
		{
			typeName: "mcs_lb_servicegroup_member", path: "/api/loadbalancing/lbservicegroupmember/", listAttr: "lb_servicegroup_members",
			page1:  map[string]interface{}{"id": "m-1", "address": "10.0.0.1", "servername": "a", "state": "UP"},
			page2:  map[string]interface{}{"id": "m-2", "address": "10.0.0.2", "servername": "b", "state": "DOWN"},
			checks: map[string]string{"state": "DOWN"},
		},
		{
			typeName: "mcs_lb_monitor", path: "/api/loadbalancing/monitor/", listAttr: "lb_monitors", byName: true,
			page1:  map[string]interface{}{"id": "mon-1", "name": "first"},
			page2:  map[string]interface{}{"id": "mon-2", "name": "target", "type": "TCP"},
			checks: map[string]string{"type": "TCP"},
		},
		{
			typeName: "mcs_cs_action", path: "/api/loadbalancing/csaction/", listAttr: "cs_actions", byName: true,
			page1:  map[string]interface{}{"id": "a-1", "name": "first"},
			page2:  map[string]interface{}{"id": "a-2", "name": "target", "lbvserver": "lbv-1"},
			checks: map[string]string{"lbvserver": "lbv-1"},
		},
		{
			typeName: "mcs_cs_policy", path: "/api/loadbalancing/cspolicy/", listAttr: "cs_policies", byName: true,
			page1:  map[string]interface{}{"id": "p-1", "name": "first"},
			page2:  map[string]interface{}{"id": "p-2", "name": "target", "application": "app-1"},
			checks: map[string]string{"application": "app-1"},
		},
		{
			typeName: "mcs_rewrite_action", path: "/api/loadbalancing/rewriteaction/", listAttr: "rewrite_actions", byName: true,
			page1:  map[string]interface{}{"id": "ra-1", "name": "first"},
			page2:  map[string]interface{}{"id": "ra-2", "name": "target", "type": "replace_all"},
			checks: map[string]string{"type": "replace_all"},
		},
		{
			typeName: "mcs_rewrite_policy", path: "/api/loadbalancing/rewritepolicy/", listAttr: "rewrite_policies", byName: true,
			page1:  map[string]interface{}{"id": "rp-1", "name": "first", "priority": 10},
			page2:  map[string]interface{}{"id": "rp-2", "name": "target", "priority": 200, "bindpoint": "RESPONSE"},
			checks: map[string]string{"priority": "200", "bindpoint": "RESPONSE"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.typeName, func(t *testing.T) {
			mock := newMockAPIServer()
			defer mock.Close()
			queries := twoPageMock(mock, tc.path, []map[string]interface{}{tc.page1}, []map[string]interface{}{tc.page2})

			listChecks := []resource.TestCheckFunc{
				resource.TestCheckResourceAttr("data."+tc.typeName+".all", tc.listAttr+".#", "2"),
				resource.TestCheckResourceAttr("data."+tc.typeName+".all", tc.listAttr+".1.id", tc.page2["id"].(string)),
			}
			for k, v := range tc.checks {
				listChecks = append(listChecks, resource.TestCheckResourceAttr("data."+tc.typeName+".all", tc.listAttr+".1."+k, v))
			}
			steps := []resource.TestStep{{
				Config: providerConfigBlock(mock.URL()) + fmt.Sprintf(`data %q "all" {}`, tc.typeName),
				Check:  resource.ComposeAggregateTestCheckFunc(listChecks...),
			}}

			if tc.byName {
				nameChecks := []resource.TestCheckFunc{
					resource.TestCheckResourceAttr("data."+tc.typeName+".one", "id", tc.page2["id"].(string)),
				}
				for k, v := range tc.checks {
					nameChecks = append(nameChecks, resource.TestCheckResourceAttr("data."+tc.typeName+".one", k, v))
				}
				steps = append(steps, resource.TestStep{
					Config: providerConfigBlock(mock.URL()) + fmt.Sprintf(`
data %q "one" {
  name = "target"
}`, tc.typeName),
					Check: resource.ComposeAggregateTestCheckFunc(nameChecks...),
				})
			}

			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
				Steps:                    steps,
			})

			for _, q := range queries() {
				if q != "" && q != "page=2" {
					t.Errorf("unexpected query string %q sent to %s", q, tc.path)
				}
			}
		})
	}
}

func TestLbSync_DnsDomainDataSource_FiltersAndNewFields(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	queries := twoPageMock(mock, "/api/dns/domains/",
		[]map[string]interface{}{{
			"uuid": "d-1", "name": "a.example", "comment": "", "provider": map[string]interface{}{"id": 7, "name": "PowerDNS"},
			"type": "external", "zone_type": "forward", "customer": "cust-1",
		}},
		[]map[string]interface{}{{
			"uuid": "d-2", "name": "2.0.192.in-addr.arpa", "comment": "", "provider": map[string]interface{}{"id": 7, "name": "PowerDNS"},
			"type": "", "zone_type": "reverse", "customer": "cust-1",
		}},
	)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_dns_domain" "test" {
  zone_type = "sideways"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_dns_domain" "test" {
  customer    = "cust-1"
  provider_id = 7
  zone_type   = "reverse"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_dns_domain.test", "domains.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_dns_domain.test", "domains.0.provider_id", "7"),
					resource.TestCheckResourceAttr("data.mcs_dns_domain.test", "domains.0.zone_type", "forward"),
					resource.TestCheckResourceAttr("data.mcs_dns_domain.test", "domains.1.zone_type", "reverse"),
					resource.TestCheckResourceAttr("data.mcs_dns_domain.test", "domains.1.provider_name", "PowerDNS"),
				),
			},
		},
	})

	q := queries()
	if len(q) == 0 || q[0] != "customer=cust-1&provider=7&zone_type=reverse" {
		t.Errorf("expected filter query on first request, got %v", q)
	}
}
