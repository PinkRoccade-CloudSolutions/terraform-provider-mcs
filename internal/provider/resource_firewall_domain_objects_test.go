package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// fwDomainStore is a stateful mock for /api/networking/domain/{domain}/{kind}/{key}/ endpoints.
// Objects are keyed by domain/kind/<keyField>; PATCH may change the key (rename).
type fwDomainStore struct {
	mu       sync.Mutex
	objects  map[string]map[string]interface{}
	defaults map[string]map[string]interface{}
	keyField map[string]string
	nextID   int
}

func newFwDomainStore(mock *mockAPIServer) *fwDomainStore {
	s := &fwDomainStore{
		objects:  map[string]map[string]interface{}{},
		defaults: map[string]map[string]interface{}{},
		keyField: map[string]string{},
		nextID:   100,
	}
	mock.On("/api/networking/domain/", s.handle)
	return s
}

func (s *fwDomainStore) kind(kind, keyField string, defaults map[string]interface{}) {
	s.keyField[kind] = keyField
	s.defaults[kind] = defaults
}

func (s *fwDomainStore) handle(w http.ResponseWriter, r *http.Request, body []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")

	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/networking/domain/"), "/"), "/")
	if len(parts) < 2 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	domain, kind := parts[0], parts[1]
	keyField := s.keyField[kind]
	var req map[string]interface{}
	if len(body) > 0 {
		_ = json.Unmarshal(body, &req)
	}

	if len(parts) == 2 && r.Method == http.MethodPost {
		obj := map[string]interface{}{"uuid": fmt.Sprintf("uuid-%d", s.nextID)}
		if keyField == "policyid" {
			obj["policyid"] = s.nextID
		}
		s.nextID++
		for k, v := range s.defaults[kind] {
			obj[k] = v
		}
		for k, v := range req {
			obj[k] = v
		}
		s.objects[fmt.Sprintf("%s/%s/%v", domain, kind, obj[keyField])] = obj
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(obj)
		return
	}
	if len(parts) != 3 {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	key := fmt.Sprintf("%s/%s/%s", domain, kind, parts[2])
	obj, ok := s.objects[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"detail":"Not found."}`)
		return
	}
	switch r.Method {
	case http.MethodGet:
		_ = json.NewEncoder(w).Encode(obj)
	case http.MethodPatch:
		for k, v := range req {
			obj[k] = v
		}
		delete(s.objects, key)
		s.objects[fmt.Sprintf("%s/%s/%v", domain, kind, obj[keyField])] = obj
		_ = json.NewEncoder(w).Encode(obj)
	case http.MethodDelete:
		delete(s.objects, key)
		w.WriteHeader(http.StatusNoContent)
	}
}

// fwFindCall returns the body of the first call with the given method and path.
func fwFindCall(mock *mockAPIServer, method, path string) (map[string]interface{}, bool) {
	for _, c := range mock.Calls() {
		if c.Method == method && c.Path == path {
			var b map[string]interface{}
			_ = json.Unmarshal([]byte(c.Body), &b)
			return b, true
		}
	}
	return nil, false
}

// fwCheckCall asserts that a call was made and that the listed body fields have the expected JSON encoding.
// A field mapped to "<absent>" must not be present in the body.
func fwCheckCall(mock *mockAPIServer, method, path string, fields map[string]string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		body, ok := fwFindCall(mock, method, path)
		if !ok {
			return fmt.Errorf("expected %s %s, calls: %v", method, path, mock.Calls())
		}
		for k, want := range fields {
			v, present := body[k]
			if want == "<absent>" {
				if present {
					return fmt.Errorf("%s %s: expected %q to be absent, got %v", method, path, k, v)
				}
				continue
			}
			got, _ := json.Marshal(v)
			if !present || string(got) != want {
				return fmt.Errorf("%s %s: expected %s=%s, got %s", method, path, k, want, got)
			}
		}
		return nil
	}
}

func TestAccFirewallObjectResource_RenameAndReplace(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	store := newFwDomainStore(mock)
	store.kind("objects", "name", map[string]interface{}{"used": true, "managed": true})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_object" "test" {
  domain  = "d1"
  name    = "old-obj"
  address = "10.0.0.1"
  subnet  = "255.255.255.255"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "managed", "true"),
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "used", "true"),
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "comment", ""),
					fwCheckCall(mock, http.MethodPost, "/api/networking/domain/d1/objects/", map[string]string{"comment": "<absent>", "used": "<absent>", "managed": "<absent>"}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_object" "test" {
  domain  = "d1"
  name    = "new-obj"
  address = "10.0.0.1"
  subnet  = "255.255.255.255"
  comment = "renamed"
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("mcs_firewall_object.test", plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "name", "new-obj"),
					resource.TestCheckResourceAttr("mcs_firewall_object.test", "comment", "renamed"),
					fwCheckCall(mock, http.MethodPatch, "/api/networking/domain/d1/objects/old-obj/", map[string]string{"name": `"new-obj"`, "comment": `"renamed"`}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_object" "test" {
  domain  = "d2"
  name    = "new-obj"
  address = "10.0.0.1"
  subnet  = "255.255.255.255"
  comment = "renamed"
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("mcs_firewall_object.test", plancheck.ResourceActionDestroyBeforeCreate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					fwCheckCall(mock, http.MethodDelete, "/api/networking/domain/d1/objects/new-obj/", nil),
					fwCheckCall(mock, http.MethodPost, "/api/networking/domain/d2/objects/", map[string]string{"name": `"new-obj"`}),
				),
			},
		},
	})
}

func TestAccFirewallObjectGroupResource_RenameAndClearMembers(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	store := newFwDomainStore(mock)
	store.kind("groups", "name", map[string]interface{}{"used": false})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_object_group" "test" {
  domain = "d1"
  name   = "grp-old"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("mcs_firewall_object_group.test", "member"),
					fwCheckCall(mock, http.MethodPost, "/api/networking/domain/d1/groups/", map[string]string{"member": "<absent>"}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_object_group" "test" {
  domain = "d1"
  name   = "grp-new"
  member = ["a", "b"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_object_group.test", "member.#", "2"),
					fwCheckCall(mock, http.MethodPatch, "/api/networking/domain/d1/groups/grp-old/", map[string]string{"name": `"grp-new"`, "member": `["a","b"]`}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_object_group" "test" {
  domain = "d1"
  name   = "grp-new"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("mcs_firewall_object_group.test", "member"),
					fwCheckCall(mock, http.MethodPatch, "/api/networking/domain/d1/groups/grp-new/", map[string]string{"member": `[]`}),
				),
			},
		},
	})
}

func TestAccFirewallServiceResource_RenameAndClearPorts(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	store := newFwDomainStore(mock)
	store.kind("services", "name", map[string]interface{}{"used": false})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_service" "test" {
  domain        = "d1"
  name          = "svc-old"
  protocol      = "TCP/UDP/SCTP"
  tcp_portrange = ["443"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "tcp_portrange.0", "443"),
					resource.TestCheckNoResourceAttr("mcs_firewall_service.test", "udp_portrange"),
					fwCheckCall(mock, http.MethodPost, "/api/networking/domain/d1/services/", map[string]string{"udp_portrange": "<absent>"}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_service" "test" {
  domain        = "d1"
  name          = "svc-new"
  protocol      = "TCP/UDP/SCTP"
  udp_portrange = ["53"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "name", "svc-new"),
					resource.TestCheckNoResourceAttr("mcs_firewall_service.test", "tcp_portrange"),
					resource.TestCheckResourceAttr("mcs_firewall_service.test", "udp_portrange.0", "53"),
					fwCheckCall(mock, http.MethodPatch, "/api/networking/domain/d1/services/svc-old/", map[string]string{
						"name": `"svc-new"`, "tcp_portrange": `[]`, "udp_portrange": `["53"]`,
					}),
				),
			},
		},
	})
}

func TestAccFirewallServiceGroupResource_RenameAndClearMembers(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	store := newFwDomainStore(mock)
	store.kind("servicegroups", "name", map[string]interface{}{"used": false})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_service_group" "test" {
  domain = "d1"
  name   = "sg-old"
  member = ["svc-1"]
}`,
				Check: resource.TestCheckResourceAttr("mcs_firewall_service_group.test", "member.#", "1"),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_service_group" "test" {
  domain = "d1"
  name   = "sg-new"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_service_group.test", "name", "sg-new"),
					resource.TestCheckNoResourceAttr("mcs_firewall_service_group.test", "member"),
					fwCheckCall(mock, http.MethodPatch, "/api/networking/domain/d1/servicegroups/sg-old/", map[string]string{"name": `"sg-new"`, "member": `[]`}),
				),
			},
		},
	})
}

func TestAccFirewallRuleResource_ComputedFieldsAndClearLists(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	store := newFwDomainStore(mock)
	store.kind("rules", "policyid", map[string]interface{}{
		"group": "grp", "origin": "mcs", "used": true, "compliant": false, "hit_count": 42,
		"last_hit": "2026-09-30T12:00:00Z", "compliancy_errors": []string{"any-service"},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_rule" "test" {
  domain  = "d1"
  enabled = true
  action  = true
  src     = ["all"]
  service = ["ALL"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "policyid", "100"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "id", "100"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "origin", "mcs"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "used", "true"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "compliant", "false"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "hit_count", "42"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "last_hit", "2026-09-30T12:00:00Z"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "compliancy_errors.0", "any-service"),
					resource.TestCheckNoResourceAttr("mcs_firewall_rule.test", "dst"),
					fwCheckCall(mock, http.MethodPost, "/api/networking/domain/d1/rules/", map[string]string{
						"policyid": "<absent>", "uuid": "<absent>", "used": "<absent>", "hit_count": "<absent>", "dst": "<absent>",
					}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_rule" "test" {
  domain  = "d1"
  enabled = false
  action  = true
  service = ["ALL"]
  comment = "tightened"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "enabled", "false"),
					resource.TestCheckResourceAttr("mcs_firewall_rule.test", "comment", "tightened"),
					resource.TestCheckNoResourceAttr("mcs_firewall_rule.test", "src"),
					fwCheckCall(mock, http.MethodPatch, "/api/networking/domain/d1/rules/100/", map[string]string{
						"src": `[]`, "enabled": "false", "policyid": "<absent>", "comment": `"tightened"`,
					}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_firewall_rule" "test" {
  domain  = "d2"
  enabled = false
  action  = true
  service = ["ALL"]
  comment = "tightened"
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("mcs_firewall_rule.test", plancheck.ResourceActionDestroyBeforeCreate)},
				},
				Check: fwCheckCall(mock, http.MethodDelete, "/api/networking/domain/d1/rules/100/", nil),
			},
		},
	})
}
