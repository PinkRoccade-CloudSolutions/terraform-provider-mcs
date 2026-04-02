package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ---------------------------------------------------------------------------
// mcs_certificate
// ---------------------------------------------------------------------------

func TestAccCertificateResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/certificate", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{"id": "cert-001"}
			for k, v := range req {
				resp[k] = v
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "cert-001", "name": "my-cert",
			})
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{"id": "cert-001"}
			for k, v := range req {
				resp[k] = v
			}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_certificate" "test" {
  name = "my-cert"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_certificate.test", "id", "cert-001"),
					resource.TestCheckResourceAttr("mcs_certificate.test", "name", "my-cert"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_contact
// ---------------------------------------------------------------------------

func TestAccContactResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/tenant/contacts", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 10, "company": req["company"], "tenant": req["tenant"],
				"firstname": req["firstname"], "lastname": req["lastname"],
				"email": req["email"], "phone": req["phone"], "address": req["address"],
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 10, "company": "ACME", "tenant": 1,
				"firstname": "John", "lastname": "Doe",
				"email": "john@example.com", "phone": "123-456", "address": "123 Main St",
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			})
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 10, "company": req["company"], "tenant": req["tenant"],
				"firstname": req["firstname"], "lastname": req["lastname"],
				"email": req["email"], "phone": req["phone"], "address": req["address"],
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_contact" "test" {
  company   = "ACME"
  firstname = "John"
  lastname  = "Doe"
  email     = "john@example.com"
  phone     = "123-456"
  address   = "123 Main St"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_contact.test", "id", "10"),
					resource.TestCheckResourceAttr("mcs_contact.test", "company", "ACME"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_customer
// ---------------------------------------------------------------------------

func TestAccCustomerResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	customerResponse := map[string]interface{}{
		"id": "cust-001", "name": "Test Customer", "contractid": "",
		"tenant": 1, "sdm": 0,
		"tech_contacts": []int{}, "admin_contacts": []int{},
		"created_at_timestamp": "2025-01-01T00:00:00Z",
		"updated_at_timestamp": "2025-01-01T00:00:00Z",
	}

	mock.On("/api/tenant/customers", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(customerResponse)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(customerResponse)
		case http.MethodPut:
			_ = json.NewEncoder(w).Encode(customerResponse)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_customer" "test" {
  name           = "Test Customer"
  contractid     = ""
  sdm            = 0
  tech_contacts  = []
  admin_contacts = []
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_customer.test", "id", "cust-001"),
					resource.TestCheckResourceAttr("mcs_customer.test", "name", "Test Customer"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_dbl
// ---------------------------------------------------------------------------

func TestAccDblResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/dbl", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 5, "ipaddress": req["ipaddress"],
				"timestamp": "2025-01-01T00:00:00Z", "occurrence": 1, "hostname": "bad-host",
			}
			if v, ok := req["source"]; ok {
				resp["source"] = v
			}
			if v, ok := req["persistent"]; ok {
				resp["persistent"] = v
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 5, "ipaddress": "10.0.0.1",
				"timestamp": "2025-01-01T00:00:00Z", "occurrence": 1, "hostname": "bad-host",
			})
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 5, "ipaddress": req["ipaddress"],
				"timestamp": "2025-01-01T00:00:00Z", "occurrence": 2, "hostname": "bad-host",
			}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_dbl" "test" {
  ipaddress = "10.0.0.1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_dbl.test", "id", "5"),
					resource.TestCheckResourceAttr("mcs_dbl.test", "ipaddress", "10.0.0.1"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_domain_dbl
// ---------------------------------------------------------------------------

func TestAccDomainDblResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/dbl/domaindbl", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 3, "domainname": req["domainname"], "source": req["source"],
				"timestamp": "2025-01-01T00:00:00Z", "occurrence": 1,
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 3, "domainname": "evil.example.com", "source": "manual",
				"timestamp": "2025-01-01T00:00:00Z", "occurrence": 1,
			})
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 3, "domainname": req["domainname"], "source": req["source"],
				"timestamp": "2025-01-01T00:00:00Z", "occurrence": 2,
			}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_domain_dbl" "test" {
  domainname = "evil.example.com"
  source     = "manual"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_domain_dbl.test", "id", "3"),
					resource.TestCheckResourceAttr("mcs_domain_dbl.test", "domainname", "evil.example.com"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_cs_action
// ---------------------------------------------------------------------------

func TestAccCsActionResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/csaction", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{"id": "act-001", "name": req["name"]}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "act-001", "name": "redirect-action",
			})
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "act-001", "name": req["name"],
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_cs_action" "test" {
  name = "redirect-action"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_cs_action.test", "id", "act-001"),
					resource.TestCheckResourceAttr("mcs_cs_action.test", "name", "redirect-action"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_cs_policy
// ---------------------------------------------------------------------------

func TestAccCsPolicyResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/cspolicy", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{"id": "pol-001", "name": req["name"]}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "pol-001", "name": "routing-policy",
			})
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "pol-001", "name": req["name"],
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_cs_policy" "test" {
  name = "routing-policy"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_cs_policy.test", "id", "pol-001"),
					resource.TestCheckResourceAttr("mcs_cs_policy.test", "name", "routing-policy"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_csv_server
// ---------------------------------------------------------------------------

func TestAccCsvServerResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	csv := map[string]interface{}{
		"id":     "csv-001",
		"name":   "",
		"ufname": "",
		"type":   "",
	}

	mock.On("/api/loadbalancing/csvserver", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			csv["name"] = req["name"]
			csv["ufname"] = req["ufname"]
			csv["type"] = req["type"]
			for _, k := range []string{"ipaddress", "policies", "certificate"} {
				if v, ok := req[k]; ok {
					csv[k] = v
				} else {
					delete(csv, k)
				}
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(csv)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(csv)
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			csv["name"] = req["name"]
			csv["ufname"] = req["ufname"]
			csv["type"] = req["type"]
			for _, k := range []string{"ipaddress", "policies", "certificate"} {
				if v, ok := req[k]; ok {
					csv[k] = v
				} else {
					delete(csv, k)
				}
			}
			_ = json.NewEncoder(w).Encode(csv)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

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
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_csv_server.test", "id", "csv-001"),
					resource.TestCheckResourceAttr("mcs_csv_server.test", "name", "my-csv"),
					resource.TestCheckNoResourceAttr("mcs_csv_server.test", "policies.#"),
					resource.TestCheckNoResourceAttr("mcs_csv_server.test", "certificate.#"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_csv_server" "test" {
  name        = "my-csv"
  ufname      = "my-csv-uf"
  type        = "HTTP"
  ipaddress   = "pip-uuid-1"
  policies    = ["pol-1"]
  certificate = ["cert-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_csv_server.test", "ipaddress", "pip-uuid-1"),
					resource.TestCheckResourceAttr("mcs_csv_server.test", "policies.#", "1"),
					resource.TestCheckResourceAttr("mcs_csv_server.test", "certificate.#", "1"),
					resource.TestCheckResourceAttr("mcs_csv_server.test", "certificate.0", "cert-1"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_firewall_object
// ---------------------------------------------------------------------------

func TestAccFirewallObjectResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	currentAddress := "10.0.0.1"
	currentSubnet := "255.255.255.255"

	mock.On("/api/networking/domain", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			if v, ok := req["address"]; ok {
				currentAddress = v.(string)
			}
			if v, ok := req["subnet"]; ok {
				currentSubnet = v.(string)
			}
			resp := map[string]interface{}{
				"name": req["name"], "uuid": "fw-obj-uuid-001", "used": false,
				"address": currentAddress, "subnet": currentSubnet,
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "test-obj", "uuid": "fw-obj-uuid-001", "used": false,
				"address": currentAddress, "subnet": currentSubnet,
			})
		case http.MethodPatch:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			if v, ok := req["address"]; ok {
				currentAddress = v.(string)
			}
			if v, ok := req["subnet"]; ok {
				currentSubnet = v.(string)
			}
			resp := map[string]interface{}{
				"name": "test-obj", "uuid": "fw-obj-uuid-001", "used": false,
				"address": currentAddress, "subnet": currentSubnet,
			}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_object" "test" {
  domain  = "test-domain"
  name    = "test-obj"
  address = "10.0.0.1"
  subnet  = "255.255.255.255"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "uuid", "fw-obj-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "name", "test-obj"),
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "address", "10.0.0.1"),
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "subnet", "255.255.255.255"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_object" "test" {
  domain  = "test-domain"
  name    = "test-obj"
  address = "10.0.0.2"
  subnet  = "255.255.255.0"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "uuid", "fw-obj-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "address", "10.0.0.2"),
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "subnet", "255.255.255.0"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_firewall_object_group
// ---------------------------------------------------------------------------

func TestAccFirewallObjectGroupResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	currentMembers := []string{"obj-1"}

	mock.On("/api/networking/domain", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"name": req["name"], "uuid": "grp-uuid-001", "used": false,
			}
			if v, ok := req["member"]; ok {
				resp["member"] = v
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "test-group", "uuid": "grp-uuid-001", "used": false,
				"member": currentMembers,
			})
		case http.MethodPatch:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"name": "test-group", "uuid": "grp-uuid-001", "used": false,
				"member": currentMembers,
			}
			if v, ok := req["member"]; ok {
				resp["member"] = v
				raw, _ := json.Marshal(v)
				_ = json.Unmarshal(raw, &currentMembers)
			}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_object_group" "test" {
  domain = "test-domain"
  name   = "test-group"
  member = ["obj-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_object_group.test", "uuid", "grp-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_object_group.test", "name", "test-group"),
					resource.TestCheckResourceAttr("mcs_firewall_object_group.test", "member.#", "1"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_object_group" "test" {
  domain = "test-domain"
  name   = "test-group"
  member = ["obj-1", "obj-2"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_object_group.test", "uuid", "grp-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_object_group.test", "member.#", "2"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_firewall_rule
// ---------------------------------------------------------------------------

func TestAccFirewallRuleResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	currentService := []string{"HTTP"}

	ruleResponse := func() map[string]interface{} {
		return map[string]interface{}{
			"enabled": true, "action": true,
			"uuid": "rule-uuid-001", "policyid": 100, "group": "",
			"src": []string{"all"}, "dst": []string{"all"},
			"src_intf": []string{"any"}, "dst_intf": []string{"any"},
			"service": currentService,
		}
	}

	mock.On("/api/networking/domain", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(ruleResponse())
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(ruleResponse())
		case http.MethodPatch:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			if v, ok := req["service"]; ok {
				raw, _ := json.Marshal(v)
				_ = json.Unmarshal(raw, &currentService)
			}
			_ = json.NewEncoder(w).Encode(ruleResponse())
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_rule" "test" {
  domain   = "test-domain"
  enabled  = true
  action   = true
  src      = ["all"]
  dst      = ["all"]
  src_intf = ["any"]
  dst_intf = ["any"]
  service  = ["HTTP"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "uuid", "rule-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "policyid", "100"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "service.0", "HTTP"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_rule" "test" {
  domain   = "test-domain"
  enabled  = true
  action   = true
  src      = ["all"]
  dst      = ["all"]
  src_intf = ["any"]
  dst_intf = ["any"]
  service  = ["HTTPS"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "uuid", "rule-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "policyid", "100"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "service.0", "HTTPS"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_firewall_service
// ---------------------------------------------------------------------------

func TestAccFirewallServiceResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	currentTcpPorts := []string{"443"}
	currentProtocol := "TCP/UDP/SCTP"

	mock.On("/api/networking/domain", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			if v, ok := req["protocol"]; ok {
				currentProtocol = v.(string)
			}
			resp := map[string]interface{}{
				"name": req["name"], "uuid": "svc-uuid-001", "used": false,
				"protocol": currentProtocol,
			}
			if v, ok := req["tcp_portrange"]; ok {
				resp["tcp_portrange"] = v
			}
			if v, ok := req["udp_portrange"]; ok {
				resp["udp_portrange"] = v
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "https-svc", "uuid": "svc-uuid-001", "used": false,
				"protocol":      currentProtocol,
				"tcp_portrange": currentTcpPorts, "udp_portrange": []string{},
			})
		case http.MethodPatch:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			if v, ok := req["protocol"]; ok {
				currentProtocol = v.(string)
			}
			resp := map[string]interface{}{
				"name": "https-svc", "uuid": "svc-uuid-001", "used": false,
				"protocol":      currentProtocol,
				"tcp_portrange": currentTcpPorts, "udp_portrange": []string{},
			}
			if v, ok := req["tcp_portrange"]; ok {
				resp["tcp_portrange"] = v
				raw, _ := json.Marshal(v)
				_ = json.Unmarshal(raw, &currentTcpPorts)
			}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_service" "test" {
  domain        = "test-domain"
  name          = "https-svc"
  protocol      = "TCP/UDP/SCTP"
  tcp_portrange = ["443"]
  udp_portrange = []
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "uuid", "svc-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "name", "https-svc"),
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "protocol", "TCP/UDP/SCTP"),
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "tcp_portrange.0", "443"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_service" "test" {
  domain        = "test-domain"
  name          = "https-svc"
  protocol      = "TCP/UDP/SCTP"
  tcp_portrange = ["443", "8443"]
  udp_portrange = []
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "uuid", "svc-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "protocol", "TCP/UDP/SCTP"),
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "tcp_portrange.#", "2"),
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "tcp_portrange.1", "8443"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_firewall_service_group
// ---------------------------------------------------------------------------

func TestAccFirewallServiceGroupResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	currentMembers := []string{"svc-1"}

	mock.On("/api/networking/domain", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"name": req["name"], "uuid": "svcgrp-uuid-001", "used": false,
			}
			if v, ok := req["member"]; ok {
				resp["member"] = v
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"name": "web-services", "uuid": "svcgrp-uuid-001", "used": false,
				"member": currentMembers,
			})
		case http.MethodPatch:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"name": "web-services", "uuid": "svcgrp-uuid-001", "used": false,
				"member": currentMembers,
			}
			if v, ok := req["member"]; ok {
				resp["member"] = v
				raw, _ := json.Marshal(v)
				_ = json.Unmarshal(raw, &currentMembers)
			}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_service_group" "test" {
  domain = "test-domain"
  name   = "web-services"
  member = ["svc-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_service_group.test", "uuid", "svcgrp-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_service_group.test", "name", "web-services"),
					resource.TestCheckResourceAttr("mcs_firewall_service_group.test", "member.#", "1"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_service_group" "test" {
  domain = "test-domain"
  name   = "web-services"
  member = ["svc-1", "svc-2"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_service_group.test", "uuid", "svcgrp-uuid-001"),
					resource.TestCheckResourceAttr("mcs_firewall_service_group.test", "member.#", "2"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_lb_monitor
// ---------------------------------------------------------------------------

func TestAccLbMonitorResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/monitor", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{"id": "mon-001", "name": req["name"]}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "mon-001", "name": "http-monitor",
			})
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "mon-001", "name": req["name"],
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_monitor" "test" {
  name = "http-monitor"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_lb_monitor.test", "id", "mon-001"),
					resource.TestCheckResourceAttr("mcs_lb_monitor.test", "name", "http-monitor"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_lb_servicegroup
// ---------------------------------------------------------------------------

func TestAccLbServicegroupResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	sg := map[string]interface{}{
		"id":   "sg-001",
		"name": "",
		"type": "",
	}

	mock.On("/api/loadbalancing/lbservicegroup", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			if _, hasMember := req["members"]; hasMember {
				t.Error("POST must not include members")
			}
			sg["name"] = req["name"]
			sg["type"] = req["type"]
			delete(sg, "members")
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(sg)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(sg)
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			sg["name"] = req["name"]
			sg["type"] = req["type"]
			if m, ok := req["members"]; ok {
				sg["members"] = m
			}
			_ = json.NewEncoder(w).Encode(sg)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup" "test" {
  name = "backend-sg"
  type = "HTTP"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_lb_servicegroup.test", "id", "sg-001"),
					resource.TestCheckResourceAttr("mcs_lb_servicegroup.test", "name", "backend-sg"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup" "test" {
  name    = "backend-sg"
  type    = "HTTP"
  members = ["member-aaa", "member-bbb"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_lb_servicegroup.test", "id", "sg-001"),
					resource.TestCheckResourceAttr("mcs_lb_servicegroup.test", "members.#", "2"),
					resource.TestCheckResourceAttr("mcs_lb_servicegroup.test", "members.0", "member-aaa"),
					resource.TestCheckResourceAttr("mcs_lb_servicegroup.test", "members.1", "member-bbb"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup" "test" {
  name    = "backend-sg"
  type    = "HTTP"
  members = ["member-ccc"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_lb_servicegroup.test", "members.#", "1"),
					resource.TestCheckResourceAttr("mcs_lb_servicegroup.test", "members.0", "member-ccc"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_lb_servicegroup_member
// ---------------------------------------------------------------------------

func TestAccLbServicegroupMemberResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/loadbalancing/lbservicegroupmember", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": "sgm-001", "address": req["address"], "servername": req["servername"],
				"port": 0,
			}
			if v, ok := req["port"]; ok {
				resp["port"] = v
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "sgm-001", "address": "10.0.0.5", "servername": "web1", "port": 0,
			})
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "sgm-001", "address": req["address"], "servername": req["servername"],
				"port": req["port"],
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lb_servicegroup_member" "test" {
  address    = "10.0.0.5"
  servername = "web1"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_lb_servicegroup_member.test", "id", "sgm-001"),
					resource.TestCheckResourceAttr("mcs_lb_servicegroup_member.test", "address", "10.0.0.5"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_lbv_server
// ---------------------------------------------------------------------------

func TestAccLbvServerResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	lbv := map[string]interface{}{
		"id":           "lbv-001",
		"name":         "",
		"servicegroup": []string{},
	}

	mock.On("/api/loadbalancing/lbvserver", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			lbv["name"] = req["name"]
			lbv["servicegroup"] = req["servicegroup"]
			if v, ok := req["ipaddress"]; ok {
				lbv["ipaddress"] = v
			} else {
				delete(lbv, "ipaddress")
			}
			if v, ok := req["certificate"]; ok {
				lbv["certificate"] = v
			} else {
				delete(lbv, "certificate")
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(lbv)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(lbv)
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			lbv["name"] = req["name"]
			lbv["servicegroup"] = req["servicegroup"]
			if v, ok := req["ipaddress"]; ok {
				lbv["ipaddress"] = v
			} else {
				delete(lbv, "ipaddress")
			}
			if v, ok := req["certificate"]; ok {
				lbv["certificate"] = v
			} else {
				delete(lbv, "certificate")
			}
			_ = json.NewEncoder(w).Encode(lbv)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lbv_server" "test" {
  name         = "my-lbv"
  servicegroup = ["sg-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_lbv_server.test", "id", "lbv-001"),
					resource.TestCheckResourceAttr("mcs_lbv_server.test", "name", "my-lbv"),
					resource.TestCheckNoResourceAttr("mcs_lbv_server.test", "certificate.#"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_lbv_server" "test" {
  name         = "my-lbv"
  ipaddress    = "pip-uuid-2"
  servicegroup = ["sg-1"]
  certificate  = ["cert-1"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_lbv_server.test", "ipaddress", "pip-uuid-2"),
					resource.TestCheckResourceAttr("mcs_lbv_server.test", "certificate.#", "1"),
					resource.TestCheckResourceAttr("mcs_lbv_server.test", "certificate.0", "cert-1"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_monitor_ip
// ---------------------------------------------------------------------------

func TestAccMonitorIPResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/dbl/monitorip", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": "mip-001", "ipaddress": req["ipaddress"], "customer": req["customer"],
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "mip-001", "ipaddress": "10.0.0.100", "customer": "acme",
			})
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "mip-001", "ipaddress": req["ipaddress"], "customer": req["customer"],
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_monitor_ip" "test" {
  ipaddress = "10.0.0.100"
  customer  = "acme"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_monitor_ip.test", "id", "mip-001"),
					resource.TestCheckResourceAttr("mcs_monitor_ip.test", "ipaddress", "10.0.0.100"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_site_to_site_vpn
// ---------------------------------------------------------------------------

func TestAccSiteToSiteVPNResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/vpn/site_to_site", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 7, "uuid": "vpn-uuid-001", "name": req["name"],
				"state": req["state"], "last_status": req["last_status"],
				"resets": req["resets"], "last_check": req["last_check"],
				"last_reset": req["last_reset"],
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": 7, "uuid": "vpn-uuid-001", "name": "hq-to-branch",
				"state": "up", "last_status": "ok", "resets": 0,
				"last_check": "2025-01-01T12:00:00Z", "last_reset": "",
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			})
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": 7, "uuid": "vpn-uuid-001", "name": req["name"],
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			}
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_site_to_site_vpn" "test" {
  name        = "hq-to-branch"
  state       = "up"
  last_status = "ok"
  resets      = 0
  last_check  = "2025-01-01T12:00:00Z"
  last_reset  = ""
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_site_to_site_vpn.test", "id", "7"),
					resource.TestCheckResourceAttr("mcs_site_to_site_vpn.test", "name", "hq-to-branch"),
					resource.TestCheckResourceAttr("mcs_site_to_site_vpn.test", "uuid", "vpn-uuid-001"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_virtual_datacenter
// ---------------------------------------------------------------------------

func TestAccVirtualDatacenterResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/virtualization/virtualdatacenter", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			resp := map[string]interface{}{
				"id": "vdc-001", "name": req["name"], "customer": req["customer"],
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "vdc-001", "name": "prod-dc", "customer": "acme",
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			})
		case http.MethodPut:
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"id": "vdc-001", "name": req["name"], "customer": req["customer"],
				"created_at_timestamp": "2025-01-01T00:00:00Z",
				"updated_at_timestamp": "2025-01-01T00:00:00Z",
			})
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_virtual_datacenter" "test" {
  name     = "prod-dc"
  customer = "acme"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_virtual_datacenter.test", "id", "vdc-001"),
					resource.TestCheckResourceAttr("mcs_virtual_datacenter.test", "name", "prod-dc"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_nat_translation
// ---------------------------------------------------------------------------

func TestAccNATTranslationResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	natResp := map[string]interface{}{
		"id": "nat-001", "public_ip": "pip-uuid-1", "interface": "if-uuid-1",
		"firewall": "fw-uuid-1", "translation": "10.0.0.1 -> 192.168.1.1",
		"private_ip": "192.168.1.1", "translation_type": "one_to_one",
		"protocol": "tcp", "customer": "acme", "description": "Web NAT",
		"state": "synced", "enabled": true,
	}

	mock.On("/api/networking/nattranslations", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(natResp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(natResp)
		case http.MethodPut:
			_ = json.NewEncoder(w).Encode(natResp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_nat_translation" "test" {
  public_ip        = "pip-uuid-1"
  interface        = "if-uuid-1"
  firewall         = "fw-uuid-1"
  translation_type = "one_to_one"
  customer         = "acme"
  description      = "Web NAT"
  protocol         = "tcp"
  enabled          = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_nat_translation.test", "id", "nat-001"),
					resource.TestCheckResourceAttr("mcs_nat_translation.test", "public_ip", "pip-uuid-1"),
					resource.TestCheckResourceAttr("mcs_nat_translation.test", "translation_type", "one_to_one"),
					resource.TestCheckResourceAttr("mcs_nat_translation.test", "enabled", "true"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_public_ip_address
// ---------------------------------------------------------------------------

func TestAccPublicIPAddressResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	pipResp := map[string]interface{}{
		"id": "pip-001", "ip_address": "203.0.113.10",
		"pool": "pool-uuid-1", "description": "Web server IP",
		"status": "available", "type": "nat", "customer": "acme",
	}

	mock.On("/api/networking/publicipaddresss", func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(pipResp)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(pipResp)
		case http.MethodPut:
			_ = json.NewEncoder(w).Encode(pipResp)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_public_ip_address" "test" {
  pool        = "pool-uuid-1"
  description = "Web server IP"
  type        = "nat"
  customer    = "acme"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_public_ip_address.test", "id", "pip-001"),
					resource.TestCheckResourceAttr("mcs_public_ip_address.test", "ip_address", "203.0.113.10"),
					resource.TestCheckResourceAttr("mcs_public_ip_address.test", "pool", "pool-uuid-1"),
					resource.TestCheckResourceAttr("mcs_public_ip_address.test", "type", "nat"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_dns_entry
// ---------------------------------------------------------------------------

func TestAccDnsEntryResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	const domainUUID = "dns-domain-uuid-1"

	createdEntry := dnsEntryAPIModel{
		Name: "www", Type: "A", Content: "192.0.2.1", Expire: 300,
	}

	mock.On("/api/dns/domains/"+domainUUID+"/entries/"+createdEntry.Name, func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
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
			var req dnsEntryAPIModel
			_ = json.Unmarshal(body, &req)
			createdEntry = req
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(createdEntry)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"count":   1,
				"next":    nil,
				"results": []dnsEntryAPIModel{createdEntry},
			})
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
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
  expire      = 300
}`, domainUUID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_dns_entry.test", "id", domainUUID+"/www/A/192.0.2.1"),
					resource.TestCheckResourceAttr("mcs_dns_entry.test", "domain_uuid", domainUUID),
					resource.TestCheckResourceAttr("mcs_dns_entry.test", "name", "www"),
					resource.TestCheckResourceAttr("mcs_dns_entry.test", "type", "A"),
					resource.TestCheckResourceAttr("mcs_dns_entry.test", "content", "192.0.2.1"),
					resource.TestCheckResourceAttr("mcs_dns_entry.test", "expire", "300"),
				),
			},
		},
	})
}
