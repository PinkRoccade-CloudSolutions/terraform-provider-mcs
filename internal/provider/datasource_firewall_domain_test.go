package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// fwPagedHandler serves pages[0] for the first page and pages[1..] for ?page=N, with DRF-style
// absolute next links. Every query seen is recorded in queries.
func fwPagedHandler(mock *mockAPIServer, path string, pages [][]map[string]interface{}, queries *[]url.Values, mu *sync.Mutex) func(http.ResponseWriter, *http.Request, []byte) {
	return func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		q := r.URL.Query()
		mu.Lock()
		*queries = append(*queries, q)
		mu.Unlock()
		page := 1
		if p := q.Get("page"); p != "" {
			_, _ = fmt.Sscanf(p, "%d", &page)
		}
		var next interface{}
		if page < len(pages) {
			nq := url.Values{}
			for k, v := range q {
				nq[k] = v
			}
			nq.Set("page", fmt.Sprint(page+1))
			next = mock.URL() + path + "?" + nq.Encode()
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":   0,
			"next":    next,
			"results": pages[page-1],
		})
	}
}

func fwCheckQueries(queries *[]url.Values, mu *sync.Mutex, key, want string, minCalls int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		mu.Lock()
		defer mu.Unlock()
		if len(*queries) < minCalls {
			return fmt.Errorf("expected at least %d list calls, got %d", minCalls, len(*queries))
		}
		for _, q := range *queries {
			if got := q.Get(key); got != want {
				return fmt.Errorf("expected query %s=%q, got %q (query %v)", key, want, got, q)
			}
		}
		return nil
	}
}

func TestAccDomainDataSource_Pagination(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var mu sync.Mutex
	var queries []url.Values
	pages := [][]map[string]interface{}{
		{{"id": 1, "uuid": "u1", "name": "prod-old", "zone": map[string]interface{}{"id": 1}, "customer": map[string]interface{}{"id": "c", "name": "C"}, "max_networks": 1}},
		{{"id": 2, "uuid": "u2", "name": "prod", "zone": map[string]interface{}{"id": 7, "name": "dmz"}, "customer": map[string]interface{}{"id": "c", "name": "C"}, "max_networks": 4}},
	}
	mock.On("/api/tenant/domains", fwPagedHandler(mock, "/api/tenant/domains/", pages, &queries, &mu))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_domain" "test" {
  name = "prod"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_domain.test", "id", "2"),
					resource.TestCheckResourceAttr("data.mcs_domain.test", "zone_id", "7"),
					resource.TestCheckResourceAttr("data.mcs_domain.test", "zone_name", "dmz"),
					fwCheckQueries(&queries, &mu, "name__icontains", "prod", 2),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_domain" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_domain.all", "domains.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_domain.all", "domains.0.zone_id", "1"),
					resource.TestCheckNoResourceAttr("data.mcs_domain.all", "domains.0.zone_name"),
					resource.TestCheckResourceAttr("data.mcs_domain.all", "domains.1.name", "prod"),
				),
			},
		},
	})
}

func TestAccFirewallDataSource_PaginationAndCustomerFilter(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var mu sync.Mutex
	var queries []url.Values
	pages := [][]map[string]interface{}{
		{{"id": "fw-1", "name": "a", "customer": "cust-9", "device": "d1", "type": "wan"}},
		{{"id": "fw-2", "name": "b", "customer": "cust-9", "device": "d2", "type": "internet", "platform": "paloalto"}},
	}
	mock.On("/api/networking/firewalls", fwPagedHandler(mock, "/api/networking/firewalls/", pages, &queries, &mu))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_firewall" "all" {
  customer = "cust-9"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_firewall.all", "firewalls.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_firewall.all", "firewalls.1.id", "fw-2"),
					resource.TestCheckResourceAttr("data.mcs_firewall.all", "firewalls.1.platform", "paloalto"),
					resource.TestCheckResourceAttr("data.mcs_firewall.all", "customer", "cust-9"),
					fwCheckQueries(&queries, &mu, "customer", "cust-9", 2),
				),
			},
		},
	})
}

func TestAccFirewallDataSource_ById(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/networking/firewalls/fw-7", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": "fw-7", "name": "edge", "customer": "c1", "customer_name": "C One", "device": "dev-7",
			"device_name": "pan-7", "platform": "paloalto", "supports_threat_protection": false, "tag_name": "t7",
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_firewall" "one" {
  id = "fw-7"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_firewall.one", "name", "edge"),
					resource.TestCheckResourceAttr("data.mcs_firewall.one", "device_name", "pan-7"),
					resource.TestCheckResourceAttr("data.mcs_firewall.one", "tag_name", "t7"),
					resource.TestCheckResourceAttr("data.mcs_firewall.one", "firewalls.#", "0"),
				),
			},
		},
	})
}

func TestAccInterfaceDataSource_Pagination(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	var mu sync.Mutex
	var queries []url.Values
	pages := [][]map[string]interface{}{
		{{"id": "if-1", "name": "eth0", "ipaddress": "10.0.0.1", "ipv6address": "", "macAddress": "aa", "vm_name": "vm1"}},
		{{"id": "if-2", "name": "eth1", "ipaddress": "10.0.0.2", "ipv6address": "", "macAddress": "bb", "vm_name": "vm2", "network": "net-2"}},
	}
	mock.On("/api/virtualization/interface", fwPagedHandler(mock, "/api/virtualization/interface/", pages, &queries, &mu))

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_interface" "one" {
  name = "eth1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_interface.one", "id", "if-2"),
					resource.TestCheckResourceAttr("data.mcs_interface.one", "network", "net-2"),
				),
			},
		},
	})
}
