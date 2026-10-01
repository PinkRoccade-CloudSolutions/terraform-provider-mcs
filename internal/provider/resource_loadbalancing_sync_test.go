package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// lbObjectMock serves a single API object at basePath: POST/PUT/PATCH merge the request body
// into the object (server-side defaults come from the initial object), GET returns it.
// Every write body is recorded, keyed by HTTP method.
type lbObjectMock struct {
	mu     sync.Mutex
	obj    map[string]interface{}
	writes []recordedWrite
	// fixed values are applied after every write, simulating server-controlled fields.
	fixed map[string]interface{}
}

type recordedWrite struct {
	Method string
	Body   map[string]interface{}
}

func newLbObjectMock(mock *mockAPIServer, basePath string, initial map[string]interface{}) *lbObjectMock {
	m := &lbObjectMock{obj: initial}
	mock.On(basePath, func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		m.mu.Lock()
		defer m.mu.Unlock()
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			m.writes = append(m.writes, recordedWrite{Method: r.Method, Body: req})
			for k, v := range req {
				m.obj[k] = v
			}
			for k, v := range m.fixed {
				m.obj[k] = v
			}
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
			}
			_ = json.NewEncoder(w).Encode(m.obj)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(m.obj)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})
	return m
}

func (m *lbObjectMock) lastWrite() map[string]interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.writes) == 0 {
		return nil
	}
	return m.writes[len(m.writes)-1].Body
}

// checkBody asserts on the most recent write body.
func (m *lbObjectMock) checkBody(f func(body map[string]interface{}) error) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		return f(m.lastWrite())
	}
}

func bodyHasEmptyList(body map[string]interface{}, key string) error {
	v, ok := body[key]
	if !ok {
		return fmt.Errorf("expected %q in request body, got %v", key, body)
	}
	l, ok := v.([]interface{})
	if !ok || len(l) != 0 {
		return fmt.Errorf("expected %q to be an empty list, got %v", key, v)
	}
	return nil
}

func bodyLacks(body map[string]interface{}, keys ...string) error {
	for _, k := range keys {
		if _, ok := body[k]; ok {
			return fmt.Errorf("expected %q to be omitted from the request body, got %v", k, body)
		}
	}
	return nil
}

func TestLbSync_CsvServerResource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	obj := newLbObjectMock(mock, "/api/loadbalancing/csvserver", map[string]interface{}{
		"id": "csv-001", "customer": "cust-001", "loadbalancer": "lb-001",
		"clientauth": false, "clientcert": "Optional",
	})

	const r = "mcs_csv_server.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_csv_server" "test" {
  name   = "my-csv"
  ufname = "my-csv-uf"
  type   = "HTTP"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_csv_server" "test" {
  name       = "my-csv"
  ufname     = "my-csv-uf"
  type       = "ssl"
  clientcert = "Sometimes"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_csv_server" "test" {
  name   = "my-csv"
  ufname = "my-csv-uf"
  type   = "ssl"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "port", "443"),
					resource.TestCheckResourceAttr(r, "customer", "cust-001"),
					resource.TestCheckResourceAttr(r, "loadbalancer", "lb-001"),
					resource.TestCheckResourceAttr(r, "clientauth", "false"),
					resource.TestCheckResourceAttr(r, "clientcert", "Optional"),
					resource.TestCheckNoResourceAttr(r, "ca_certificate.#"),
					obj.checkBody(func(b map[string]interface{}) error {
						if b["port"] != float64(443) {
							return fmt.Errorf("expected default port 443 in body, got %v", b["port"])
						}
						return bodyLacks(b, "ca_certificate", "certificate", "policies", "customer", "clientauth", "clientcert")
					}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_csv_server" "test" {
  name           = "my-csv"
  ufname         = "my-csv-uf"
  type           = "ssl"
  port           = 8443
  certificate    = ["cert-1"]
  ca_certificate = ["ca-1", "ca-2"]
  clientauth     = true
  clientcert     = "Mandatory"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "port", "8443"),
					resource.TestCheckResourceAttr(r, "ca_certificate.#", "2"),
					resource.TestCheckResourceAttr(r, "ca_certificate.1", "ca-2"),
					resource.TestCheckResourceAttr(r, "clientauth", "true"),
					resource.TestCheckResourceAttr(r, "clientcert", "Mandatory"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_csv_server" "test" {
  name       = "my-csv"
  ufname     = "my-csv-uf"
  type       = "ssl"
  port       = 8443
  clientauth = true
  clientcert = "Mandatory"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(r, "ca_certificate.#"),
					resource.TestCheckNoResourceAttr(r, "certificate.#"),
					obj.checkBody(func(b map[string]interface{}) error {
						if err := bodyHasEmptyList(b, "ca_certificate"); err != nil {
							return err
						}
						return bodyHasEmptyList(b, "certificate")
					}),
				),
			},
		},
	})
}

func TestLbSync_LbvServerResource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	obj := newLbObjectMock(mock, "/api/loadbalancing/lbvserver", map[string]interface{}{
		"id": "lbv-001", "port": nil, "customer": "cust-001", "loadbalancer": "lb-001",
	})

	const r = "mcs_lbv_server.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lbv_server" "test" {
  name         = "my-lbv"
  type         = "SSL"
  servicegroup = ["sg-1"]
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				PreConfig: func() {
					// Simulate the API returning a null port for a non routed loadbalancer.
					obj.mu.Lock()
					obj.fixed = map[string]interface{}{"port": nil}
					obj.mu.Unlock()
				},
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lbv_server" "test" {
  name         = "my-lbv"
  servicegroup = ["sg-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "type", "ssl"),
					resource.TestCheckResourceAttr(r, "port", "0"),
					resource.TestCheckResourceAttr(r, "customer", "cust-001"),
					resource.TestCheckNoResourceAttr(r, "ca_certificate.#"),
					obj.checkBody(func(b map[string]interface{}) error {
						if b["type"] != "ssl" {
							return fmt.Errorf("expected default type ssl in body, got %v", b["type"])
						}
						if v, ok := b["ipaddress"]; !ok || v != nil {
							return fmt.Errorf("expected explicit null ipaddress, got %v", b)
						}
						return bodyLacks(b, "ca_certificate", "certificate")
					}),
				),
			},
			{
				PreConfig: func() {
					obj.mu.Lock()
					obj.fixed = nil
					obj.mu.Unlock()
				},
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lbv_server" "test" {
  name           = "my-lbv"
  type           = "tcp"
  port           = 443
  ipaddress      = "pip-1"
  servicegroup   = ["sg-1"]
  ca_certificate = ["ca-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "type", "tcp"),
					resource.TestCheckResourceAttr(r, "port", "443"),
					resource.TestCheckResourceAttr(r, "ipaddress", "pip-1"),
					resource.TestCheckResourceAttr(r, "ca_certificate.#", "1"),
					resource.TestCheckResourceAttr(r, "ca_certificate.0", "ca-1"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lbv_server" "test" {
  name         = "my-lbv"
  type         = "tcp"
  port         = 443
  servicegroup = ["sg-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(r, "ca_certificate.#"),
					resource.TestCheckNoResourceAttr(r, "ipaddress"),
					obj.checkBody(func(b map[string]interface{}) error {
						if v, ok := b["ipaddress"]; !ok || v != nil {
							return fmt.Errorf("expected explicit null ipaddress to clear it, got %v", b)
						}
						return bodyHasEmptyList(b, "ca_certificate")
					}),
				),
			},
		},
	})
}

func TestLbSync_LbServicegroupResource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	obj := newLbObjectMock(mock, "/api/loadbalancing/lbservicegroup", map[string]interface{}{
		"id": "sg-001", "state": "enable", "healthmonitor": "YES", "cip": "DISABLED",
		"cipheader": "X-Forwarded-For", "client_certificate": nil,
		"customer": "cust-001", "loadbalancer": "lb-001",
	})

	const r = "mcs_lb_servicegroup.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup" "test" {
  name = "backend-sg"
  type = "http"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup" "test" {
  name = "backend-sg"
  type = "HTTP"
  cip  = "on"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup" "test" {
  name     = "backend-sg"
  type     = "HTTP"
  members  = ["member-a"]
  monitors = ["mon-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "state", "enable"),
					resource.TestCheckResourceAttr(r, "healthmonitor", "YES"),
					resource.TestCheckResourceAttr(r, "cip", "DISABLED"),
					resource.TestCheckResourceAttr(r, "cipheader", "X-Forwarded-For"),
					resource.TestCheckNoResourceAttr(r, "client_certificate"),
					resource.TestCheckResourceAttr(r, "members.#", "1"),
					resource.TestCheckResourceAttr(r, "monitors.#", "1"),
					resource.TestCheckResourceAttr(r, "monitors.0", "mon-1"),
					func(_ *terraform.State) error {
						obj.mu.Lock()
						defer obj.mu.Unlock()
						if len(obj.writes) != 2 || obj.writes[0].Method != http.MethodPost || obj.writes[1].Method != http.MethodPut {
							return fmt.Errorf("expected POST then PUT, got %d writes", len(obj.writes))
						}
						if _, ok := obj.writes[0].Body["members"]; ok {
							return fmt.Errorf("POST must not include members: %v", obj.writes[0].Body)
						}
						if _, ok := obj.writes[0].Body["monitors"]; !ok {
							return fmt.Errorf("POST must include monitors: %v", obj.writes[0].Body)
						}
						return nil
					},
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup" "test" {
  name               = "backend-sg"
  type               = "SSL"
  state              = "disable"
  healthmonitor      = "NO"
  members            = ["member-a"]
  client_certificate = "cert-1"
  cip                = "ENABLED"
  cipheader          = "X-Real-IP"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "type", "SSL"),
					resource.TestCheckResourceAttr(r, "state", "disable"),
					resource.TestCheckResourceAttr(r, "healthmonitor", "NO"),
					resource.TestCheckResourceAttr(r, "client_certificate", "cert-1"),
					resource.TestCheckResourceAttr(r, "cip", "ENABLED"),
					resource.TestCheckResourceAttr(r, "cipheader", "X-Real-IP"),
					resource.TestCheckNoResourceAttr(r, "monitors.#"),
					obj.checkBody(func(b map[string]interface{}) error {
						return bodyHasEmptyList(b, "monitors")
					}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup" "test" {
  name      = "backend-sg"
  type      = "SSL"
  state     = "disable"
  cip       = "ENABLED"
  cipheader = "X-Real-IP"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(r, "client_certificate"),
					resource.TestCheckNoResourceAttr(r, "members.#"),
					obj.checkBody(func(b map[string]interface{}) error {
						if v, ok := b["client_certificate"]; !ok || v != nil {
							return fmt.Errorf("expected explicit null client_certificate, got %v", b)
						}
						return bodyHasEmptyList(b, "members")
					}),
				),
			},
		},
	})
}

func TestLbSync_LbServicegroupMemberResource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	newLbObjectMock(mock, "/api/loadbalancing/lbservicegroupmember", map[string]interface{}{
		"id": "sgm-001", "state": "UP", "weight": 1, "customer": "cust-001", "loadbalancer": "lb-001",
	})

	const r = "mcs_lb_servicegroup_member.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup_member" "test" {
  address    = "10.0.0.5"
  servername = "web1"
  state      = "down"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup_member" "test" {
  address    = "10.0.0.5"
  servername = "web1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "state", "UP"),
					resource.TestCheckResourceAttr(r, "weight", "1"),
					resource.TestCheckResourceAttr(r, "port", "0"),
					resource.TestCheckResourceAttr(r, "customer", "cust-001"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup_member" "test" {
  address    = "10.0.0.5"
  servername = "web1"
  port       = 8080
  state      = "DOWN"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "state", "DOWN"),
					resource.TestCheckResourceAttr(r, "port", "8080"),
				),
			},
		},
	})
}

func TestLbSync_LbMonitorResource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	newLbObjectMock(mock, "/api/loadbalancing/monitor", map[string]interface{}{
		"id": "mon-001", "type": "HTTP", "interval": 5, "resptimeout": 2, "downtime": 30,
		"respcode": `["200"]`, "secure": "", "httprequest": "", "protected": false,
		"customer": "cust-001", "loadbalancer": "lb-001",
	})

	const r = "mcs_lb_monitor.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_monitor" "test" {
  name = "http-monitor"
  type = "icmp"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_monitor" "test" {
  name   = "http-monitor"
  secure = "yes"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_monitor" "test" {
  name = "http-monitor"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "type", "HTTP"),
					resource.TestCheckResourceAttr(r, "interval", "5"),
					resource.TestCheckResourceAttr(r, "respcode", `["200"]`),
					resource.TestCheckResourceAttr(r, "secure", ""),
					resource.TestCheckResourceAttr(r, "protected", "false"),
					resource.TestCheckResourceAttr(r, "customer", "cust-001"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_monitor" "test" {
  name     = "http-monitor"
  type     = "tcp"
  respcode = "[\"200\", \"403\", \"500\"]"
  secure   = "YES"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "type", "tcp"),
					resource.TestCheckResourceAttr(r, "respcode", `["200", "403", "500"]`),
					resource.TestCheckResourceAttr(r, "secure", "YES"),
				),
			},
		},
	})
}

func TestLbSync_RewritePolicyResource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	obj := newLbObjectMock(mock, "/api/loadbalancing/rewritepolicy", map[string]interface{}{
		"id": "rw-pol-001", "rule": "", "action": nil, "undefaction": "", "comment": "", "priority": 100,
		"bindpoint": "REQUEST", "gotopriorityexpression": "END", "customer": "cust-001", "loadbalancer": "lb-001",
	})

	const r = "mcs_rewrite_policy.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_rewrite_policy" "test" {
  name      = "rw"
  bindpoint = "BOTH"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_rewrite_policy" "test" {
  name                   = "rw"
  gotopriorityexpression = "next"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_rewrite_policy" "test" {
  name     = "rw"
  priority = 1.5
}`,
				ExpectError: regexpMustCompile(`(?s)Inappropriate value|integer|whole number`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_rewrite_policy" "test" {
  name = "rw"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "priority", "100"),
					resource.TestCheckResourceAttr(r, "bindpoint", "REQUEST"),
					resource.TestCheckResourceAttr(r, "gotopriorityexpression", "END"),
					resource.TestCheckResourceAttr(r, "comment", ""),
					resource.TestCheckNoResourceAttr(r, "action"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_rewrite_policy" "test" {
  name                   = "rw"
  action                 = "rw-act-1"
  priority               = 250
  bindpoint              = "RESPONSE"
  gotopriorityexpression = "USE_INVOCATION_RESULT"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "priority", "250"),
					resource.TestCheckResourceAttr(r, "action", "rw-act-1"),
					resource.TestCheckResourceAttr(r, "bindpoint", "RESPONSE"),
					resource.TestCheckResourceAttr(r, "gotopriorityexpression", "USE_INVOCATION_RESULT"),
					obj.checkBody(func(b map[string]interface{}) error {
						// priority must be serialised as an integer.
						raw, _ := json.Marshal(b["priority"])
						if string(raw) != "250" {
							return fmt.Errorf("expected integer priority 250, got %s", raw)
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestLbSync_RewriteActionResource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	// The API returns blank strings for unset text fields.
	newLbObjectMock(mock, "/api/loadbalancing/rewriteaction", map[string]interface{}{
		"id": "rw-act-001", "type": "replace", "target": "", "stringbuilderexpr": "", "search": "",
		"comment": "", "customer": "cust-001", "loadbalancer": "lb-001",
	})

	const r = "mcs_rewrite_action.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_rewrite_action" "test" {
  name = "rwa"
  type = "REPLACE"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_rewrite_action" "test" {
  name = "rwa"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(r, "type", "replace"),
					resource.TestCheckResourceAttr(r, "target", ""),
					resource.TestCheckResourceAttr(r, "customer", "cust-001"),
					resource.TestCheckResourceAttr(r, "loadbalancer", "lb-001"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_rewrite_action" "test" {
  name = "rwa"
  type = "insert_http_header"
}`,
				Check: resource.TestCheckResourceAttr(r, "type", "insert_http_header"),
			},
		},
	})
}

func TestLbSync_CsActionAndPolicyServerFilledFields(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	action := newLbObjectMock(mock, "/api/loadbalancing/csaction", map[string]interface{}{
		"id": "act-001", "lbvserver": nil, "customer": "cust-001", "loadbalancer": "lb-001",
	})
	newLbObjectMock(mock, "/api/loadbalancing/cspolicy", map[string]interface{}{
		"id": "pol-001", "action": nil, "expression": "", "application": nil,
		"customer": "cust-001", "loadbalancer": "lb-001",
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_cs_action" "test" {
  name      = "act"
  lbvserver = "lbv-1"
}

resource "mcs_cs_policy" "test" {
  name   = "pol"
  action = mcs_cs_action.test.id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_cs_action.test", "lbvserver", "lbv-1"),
					resource.TestCheckResourceAttr("mcs_cs_action.test", "customer", "cust-001"),
					resource.TestCheckResourceAttr("mcs_cs_action.test", "loadbalancer", "lb-001"),
					resource.TestCheckResourceAttr("mcs_cs_policy.test", "action", "act-001"),
					resource.TestCheckResourceAttr("mcs_cs_policy.test", "expression", ""),
					resource.TestCheckResourceAttr("mcs_cs_policy.test", "customer", "cust-001"),
					resource.TestCheckNoResourceAttr("mcs_cs_policy.test", "application"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_cs_action" "test" {
  name = "act"
}

resource "mcs_cs_policy" "test" {
  name   = "pol"
  action = mcs_cs_action.test.id
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("mcs_cs_action.test", "lbvserver"),
					action.checkBody(func(b map[string]interface{}) error {
						if v, ok := b["lbvserver"]; !ok || v != nil {
							return fmt.Errorf("expected explicit null lbvserver, got %v", b)
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestLbSync_DnsEntryResource_BareArrayAndExpireValidation(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	const domainUUID = "dns-domain-uuid-2"
	var entry dnsEntryAPIModel

	mock.On("/api/dns/domains/"+domainUUID+"/entries/www", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})
	mock.On("/api/dns/domains/"+domainUUID+"/entries", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			_ = json.Unmarshal(body, &entry)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(entry)
		case http.MethodGet:
			// The spec documents a bare array for this endpoint.
			_ = json.NewEncoder(w).Encode([]dnsEntryAPIModel{
				{Name: "other", Type: "A", Content: "192.0.2.9", Expire: 60},
				entry,
			})
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + fmt.Sprintf(`
resource "mcs_dns_entry" "test" {
  domain_uuid = %q
  name        = "www"
  type        = "A"
  content     = "192.0.2.1"
  expire      = 30
}`, domainUUID),
				ExpectError: regexpMustCompile(`value must be between 60 and 604800`),
			},
			{
				Config: providerConfigBlock(mock.URL()) + fmt.Sprintf(`
resource "mcs_dns_entry" "test" {
  domain_uuid = %q
  name        = "www"
  type        = "A"
  content     = "192.0.2.1"
  expire      = 3600
}`, domainUUID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_dns_entry.test", "id", domainUUID+"/www/A/192.0.2.1"),
					resource.TestCheckResourceAttr("mcs_dns_entry.test", "expire", "3600"),
				),
			},
		},
	})

	for _, c := range mock.Calls() {
		if c.Method == http.MethodDelete && !strings.HasSuffix(c.Path, "/entries/www/") {
			t.Errorf("unexpected delete path %s", c.Path)
		}
	}
}
