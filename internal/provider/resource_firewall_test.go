package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

// fwFirewallMock serves /api/networking/firewalls/ statefully; created firewalls get ids fw-1, fw-2, ...
func fwFirewallMock(mock *mockAPIServer) {
	var mu sync.Mutex
	objects := map[string]map[string]interface{}{}
	next := 1
	defaults := map[string]interface{}{
		"name": "", "description": "", "type": "internet", "context": "root",
		"external_interface": "", "internal_interface": "", "default_log_profile": "",
		"default_protect_profile": "", "multi_tenant": false, "tag_name": "", "nat_ip_sync_enabled": false,
		"customer_name": "Acme", "device_name": "fmg-01", "platform": "fortinet", "supports_threat_protection": true,
	}

	mock.On("/api/networking/firewalls/", func(w http.ResponseWriter, r *http.Request, body []byte) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)

		id := r.URL.Path[len("/api/networking/firewalls/"):]
		if len(id) > 0 && id[len(id)-1] == '/' {
			id = id[:len(id)-1]
		}
		if id == "" && r.Method == http.MethodPost {
			obj := map[string]interface{}{}
			for k, v := range defaults {
				obj[k] = v
			}
			for k, v := range req {
				obj[k] = v
			}
			obj["id"] = fmt.Sprintf("fw-%d", next)
			next++
			objects[obj["id"].(string)] = obj
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(obj)
			return
		}
		obj, ok := objects[id]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"detail":"Not found."}`))
			return
		}
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(obj)
		case http.MethodPatch:
			for k, v := range req {
				obj[k] = v
			}
			_ = json.NewEncoder(w).Encode(obj)
		case http.MethodDelete:
			delete(objects, id)
			w.WriteHeader(http.StatusNoContent)
		}
	})
}

func TestAccFirewallResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	fwFirewallMock(mock)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall" "test" {
  customer = "cust-1"
  device   = "dev-1"
  name     = "edge"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall.test", "id", "fw-1"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "name", "edge"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "type", "internet"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "context", "root"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "multi_tenant", "false"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "customer_name", "Acme"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "device_name", "fmg-01"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "platform", "fortinet"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "supports_threat_protection", "true"),
					fwCheckCall(mock, http.MethodPost, "/api/networking/firewalls/", map[string]string{
						"customer": `"cust-1"`, "device": `"dev-1"`, "name": `"edge"`,
						"type": "<absent>", "multi_tenant": "<absent>", "customer_name": "<absent>", "id": "<absent>",
					}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall" "test" {
  customer                = "cust-1"
  device                  = "dev-1"
  name                    = "edge"
  description             = "Edge firewall"
  type                    = "wan"
  context                 = "vdom-b"
  external_interface      = "port1"
  internal_interface      = "port2"
  default_log_profile     = "log"
  default_protect_profile = "protect"
  multi_tenant            = true
  tag_name                = "tenant-a"
  nat_ip_sync_enabled     = true
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("mcs_firewall.test", plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall.test", "id", "fw-1"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "description", "Edge firewall"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "type", "wan"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "context", "vdom-b"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "external_interface", "port1"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "internal_interface", "port2"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "default_log_profile", "log"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "default_protect_profile", "protect"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "multi_tenant", "true"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "tag_name", "tenant-a"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "nat_ip_sync_enabled", "true"),
					fwCheckCall(mock, http.MethodPatch, "/api/networking/firewalls/fw-1/", map[string]string{
						"type": `"wan"`, "multi_tenant": "true", "nat_ip_sync_enabled": "true", "tag_name": `"tenant-a"`,
					}),
				),
			},
			{
				ResourceName:      "mcs_firewall.test",
				ImportState:       true,
				ImportStateId:     "fw-1",
				ImportStateVerify: true,
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall" "test" {
  customer = "cust-1"
  device   = "dev-2"
  name     = "edge"
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("mcs_firewall.test", plancheck.ResourceActionDestroyBeforeCreate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall.test", "id", "fw-2"),
					resource.TestCheckResourceAttr("mcs_firewall.test", "device", "dev-2"),
					fwCheckCall(mock, http.MethodDelete, "/api/networking/firewalls/fw-1/", nil),
				),
			},
		},
	})
}

func TestAccFirewallResource_InvalidType(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories("http://127.0.0.1:1"),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock("http://127.0.0.1:1") + `
resource "mcs_firewall" "test" {
  customer = "c"
  device   = "d"
  type     = "dmz"
}`,
				PlanOnly:    true,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
		},
	})
}
