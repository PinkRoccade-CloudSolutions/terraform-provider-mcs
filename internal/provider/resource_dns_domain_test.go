package provider

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func newDnsDomainMock(mock *mockAPIServer) (*sync.Mutex, map[string]interface{}, *[]recordedWrite) {
	var mu sync.Mutex
	domain := map[string]interface{}{
		"uuid": "dns-uuid-2", "name": "example.com", "comment": "synced", "enddate": nil,
		"customer": "cust-1", "provider": map[string]interface{}{"id": 3, "name": "PowerDNS"},
		"type": "external", "zone_type": "forward",
	}
	var writes []recordedWrite

	mock.On("/api/dns/domains/", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		mu.Lock()
		defer mu.Unlock()

		if r.URL.Path == "/api/dns/domains/" {
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			other := map[string]interface{}{
				"uuid": "dns-uuid-1", "name": "sub.example.com", "comment": "", "enddate": nil, "customer": nil,
				"provider": map[string]interface{}{"id": 3, "name": "PowerDNS"}, "type": "", "zone_type": "forward",
			}
			if r.URL.Query().Get("page") == "2" {
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"count": 2, "next": nil, "results": []interface{}{domain}})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"count": 2, "next": mock.URL() + "/api/dns/domains/?name__icontains=example.com&page=2", "results": []interface{}{other},
			})
			return
		}

		if !strings.HasPrefix(r.URL.Path, "/api/dns/domains/dns-uuid-2/") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(domain)
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			writes = append(writes, recordedWrite{Method: r.Method, Body: req})
			for k, v := range req {
				domain[k] = v
			}
			_ = json.NewEncoder(w).Encode(domain)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	return &mu, domain, &writes
}

func TestAccDnsDomainResource_AdoptUpdateForget(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	mu, _, writes := newDnsDomainMock(mock)

	const r = "mcs_dns_domain.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_dns_domain" "test" {
  name = "example.com"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "id", "dns-uuid-2"),
					resource.TestCheckResourceAttr(r, "comment", "synced"),
					resource.TestCheckResourceAttr(r, "customer", "cust-1"),
					resource.TestCheckResourceAttr(r, "type", "external"),
					resource.TestCheckResourceAttr(r, "zone_type", "forward"),
					resource.TestCheckResourceAttr(r, "provider_id", "3"),
					resource.TestCheckResourceAttr(r, "provider_name", "PowerDNS"),
					resource.TestCheckNoResourceAttr(r, "enddate"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_dns_domain" "test" {
  name      = "example.com"
  comment   = "managed by terraform"
  enddate   = "2027-06-30"
  type      = "internal"
  zone_type = "forward"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "id", "dns-uuid-2"),
					resource.TestCheckResourceAttr(r, "comment", "managed by terraform"),
					resource.TestCheckResourceAttr(r, "enddate", "2027-06-30"),
					resource.TestCheckResourceAttr(r, "type", "internal"),
				),
			},
			{
				ResourceName:      r,
				ImportState:       true,
				ImportStateId:     "dns-uuid-2",
				ImportStateVerify: true,
			},
		},
	})

	mu.Lock()
	defer mu.Unlock()
	if len(*writes) != 2 {
		t.Fatalf("expected 2 PUTs (adopt + update), got %d", len(*writes))
	}
	first := (*writes)[0].Body
	if first["name"] != "example.com" {
		t.Errorf("adopt PUT must send the name, got %v", first)
	}
	for _, k := range []string{"comment", "enddate", "customer", "type", "zone_type"} {
		if _, ok := first[k]; ok {
			t.Errorf("adopt PUT must not send unconfigured %q, got %v", k, first)
		}
	}
	if (*writes)[1].Body["comment"] != "managed by terraform" {
		t.Errorf("update PUT missing comment: %v", (*writes)[1].Body)
	}
	for _, c := range mock.Calls() {
		if c.Method == http.MethodDelete || c.Method == http.MethodPost {
			t.Errorf("unexpected %s %s: dns domains must never be created or deleted", c.Method, c.Path)
		}
	}
}

func TestAccDnsDomainResource_NotFound(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	newDnsDomainMock(mock)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_dns_domain" "test" {
  name = "example"
}`,
				ExpectError: regexpMustCompile(`DNS domain not found`),
			},
		},
	})
}

func TestAccDnsDomainResource_Validation(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	newDnsDomainMock(mock)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_dns_domain" "test" {
  name      = "example.com"
  zone_type = "both"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_dns_domain" "test" {
  name = "example.com"
  type = "public"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
		},
	})
}
