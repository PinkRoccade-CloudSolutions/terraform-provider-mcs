package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func init() {
	networkOperationPollInterval = 10 * time.Millisecond
}

type networkMockJob struct {
	id        int
	undeploy  bool
	remaining int // polls answered as pending before the job finishes
	networkID string
}

// networkV3Mock serves the v3 networking API for one tenant: deploys and teardowns are queued
// as jobs that report pending for pendingPolls polls before finishing.
type networkV3Mock struct {
	mu           sync.Mutex
	networks     map[string]map[string]interface{}
	jobs         map[int]*networkMockJob
	nextJob      int
	nextNet      int
	pendingPolls int
	deployFails  bool // the deploy writes the network row, then the job fails
	neverFinish  bool // deploys stay pending forever
	postStatus   int  // when set, POST answers with this status instead of queueing a deploy
	conflictOnce bool // the next DELETE answers 409 and the network disappears after a few GETs
	goneAfter    int
	lastPostBody map[string]interface{}
}

func newNetworkV3Mock(mock *mockAPIServer) *networkV3Mock {
	m := &networkV3Mock{
		networks:     map[string]map[string]interface{}{},
		jobs:         map[int]*networkMockJob{},
		nextJob:      100,
		pendingPolls: 2,
	}
	mock.On("/api/v3/networking/operations/", m.handleOperation)
	mock.On("/api/v3/networking/networks/", m.handleNetworks)
	mock.On("/api/tenant/domains/", func(w http.ResponseWriter, _ *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"results": []interface{}{
			map[string]interface{}{"id": 4, "uuid": "dom-uuid-4", "name": "test"},
			map[string]interface{}{"id": 5, "uuid": "dom-uuid-5", "name": "prod"},
		}})
	})
	return m
}

func (m *networkV3Mock) addNetwork(id string, description string, prefix string) map[string]interface{} {
	n := map[string]interface{}{
		"id": id, "name": "NET-" + id, "domain": 5,
		"domain_detail": map[string]interface{}{"id": 5, "vdom_id": "vd5", "name": "prod"},
		"customer":      "cust-1", "customer_name": "Customer One", "description": description,
		"type": "vlan", "parent": "LACP Edge netwe", "vlanid": 120,
		"ipv4_prefix": prefix, "ipv4_address": "10.20.0.1", "ipv6_prefix": "", "ipv6_address": "",
		"dhcp_server": false, "dhcp_server_address_pool": "", "dhcp_server_netmask": "",
		"dhcp_server_dns_servers": "", "dhcp_server_lease_time": 86400, "visible": true,
		"created_at_timestamp": "2026-10-01T10:00:00Z", "updated_at_timestamp": "2026-10-01T10:00:00Z",
	}
	m.networks[id] = n
	return n
}

func (m *networkV3Mock) jobJSON(j *networkMockJob, status, result string, message interface{}) map[string]interface{} {
	name := "deploy_network"
	if j.undeploy {
		name = "undeploy_network"
	}
	var network interface{}
	if n, ok := m.networks[j.networkID]; ok && j.networkID != "" {
		network = n
	}
	return map[string]interface{}{
		"job_id": j.id, "jobname": name, "message": message, "network": network,
		"status": map[string]interface{}{"status": status}, "result": map[string]interface{}{"result": result},
	}
}

func (m *networkV3Mock) handleOperation(w http.ResponseWriter, r *http.Request, _ []byte) {
	w.Header().Set("Content-Type", "application/json")
	m.mu.Lock()
	defer m.mu.Unlock()

	var id int
	_, _ = fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/api/v3/networking/operations/"), "%d", &id)
	j, ok := m.jobs[id]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if (m.neverFinish && !j.undeploy) || j.remaining > 0 {
		j.remaining--
		_ = json.NewEncoder(w).Encode(m.jobJSON(j, "RUNNING", "PENDING", nil))
		return
	}
	if j.undeploy {
		delete(m.networks, j.networkID)
		_ = json.NewEncoder(w).Encode(m.jobJSON(j, "PROCESSED", "SUCCESS", "Network removed"))
		return
	}
	if m.deployFails {
		_ = json.NewEncoder(w).Encode(m.jobJSON(j, "PROCESSED", "EXCEPTION", "FortiGate rejected the interface"))
		return
	}
	_ = json.NewEncoder(w).Encode(m.jobJSON(j, "PROCESSED", "SUCCESS", "Network deployed"))
}

func (m *networkV3Mock) handleNetworks(w http.ResponseWriter, r *http.Request, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	m.mu.Lock()
	defer m.mu.Unlock()

	id := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v3/networking/networks/"), "/")
	if id == "" {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if m.postStatus != 0 {
			w.WriteHeader(m.postStatus)
			_, _ = fmt.Fprint(w, `{"detail":"boom"}`)
			return
		}
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)
		m.lastPostBody = req

		m.nextNet++
		netID := fmt.Sprintf("net-uuid-%d", m.nextNet)
		prefix := "10.20.0.0"
		if p, ok := req["prefix"].(string); ok {
			prefix = p
		}
		m.addNetwork(netID, fmt.Sprint(req["description"]), fmt.Sprintf("%s/%v", prefix, req["bitmask"]))

		m.nextJob++
		j := &networkMockJob{id: m.nextJob, remaining: m.pendingPolls, networkID: netID}
		m.jobs[j.id] = j
		w.WriteHeader(http.StatusAccepted)
		resp := m.jobJSON(j, "QUEUED", "PENDING", nil)
		resp["network"] = nil
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	n, ok := m.networks[id]
	switch r.Method {
	case http.MethodGet:
		if ok && m.goneAfter > 0 {
			m.goneAfter--
			if m.goneAfter == 0 {
				delete(m.networks, id)
				ok = false
			}
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"detail":"Not found."}`)
			return
		}
		_ = json.NewEncoder(w).Encode(n)
	case http.MethodDelete:
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if m.conflictOnce {
			m.conflictOnce = false
			m.goneAfter = 3
			w.WriteHeader(http.StatusConflict)
			_, _ = fmt.Fprint(w, `{"detail":"A teardown is already running"}`)
			return
		}
		m.nextJob++
		j := &networkMockJob{id: m.nextJob, undeploy: true, remaining: m.pendingPolls, networkID: id}
		m.jobs[j.id] = j
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(m.jobJSON(j, "QUEUED", "PENDING", nil))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *networkV3Mock) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.networks)
}

func countCalls(mock *mockAPIServer, method, prefix string) int {
	n := 0
	for _, c := range mock.Calls() {
		if c.Method == method && strings.HasPrefix(c.Path, prefix) {
			n++
		}
	}
	return n
}

func networkResourceConfig(url, extra string) string {
	return providerConfigBlock(url) + fmt.Sprintf(`
resource "mcs_network" "test" {
  domain_uuid     = "dom-uuid-5"
  product         = "DMZ"
  bitmask         = 26
  network_pool_id = "pool-uuid-1"
%s
}`, extra)
}

const networkDescription = `  description     = "web tier"`

func TestAccNetworkResource_CreateImportDestroy(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)

	addr := "mcs_network.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		CheckDestroy: func(_ *terraform.State) error {
			if n := nm.count(); n != 0 {
				return fmt.Errorf("expected all networks to be removed, %d left", n)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: networkResourceConfig(mock.URL(), networkDescription),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", "net-uuid-1"),
					resource.TestCheckResourceAttr(addr, "name", "NET-net-uuid-1"),
					resource.TestCheckResourceAttr(addr, "description", "web tier"),
					resource.TestCheckResourceAttr(addr, "domain_id", "5"),
					resource.TestCheckResourceAttr(addr, "domain_name", "prod"),
					resource.TestCheckResourceAttr(addr, "domain_vdom_id", "vd5"),
					resource.TestCheckResourceAttr(addr, "vlan_id", "120"),
					resource.TestCheckResourceAttr(addr, "ipv4_prefix", "10.20.0.0/26"),
					resource.TestCheckResourceAttr(addr, "type", "vlan"),
					resource.TestCheckResourceAttr(addr, "dhcp_server", "false"),
					resource.TestCheckResourceAttr(addr, "dhcp_server_lease_time", "86400"),
					resource.TestCheckResourceAttr(addr, "visible", "true"),
					resource.TestCheckResourceAttr(addr, "customer_name", "Customer One"),
					resource.TestCheckNoResourceAttr(addr, "prefix"),
				),
			},
			{
				Config:            networkResourceConfig(mock.URL(), networkDescription),
				ResourceName:      addr,
				ImportState:       true,
				ImportStateVerify: true,
				// The API never returns these create-only arguments.
				ImportStateVerifyIgnore: []string{"product", "network_pool_id", "prefix"},
			},
		},
	})

	if got := countCalls(mock, http.MethodPost, "/api/v3/networking/networks/"); got != 1 {
		t.Errorf("expected 1 deploy POST, got %d", got)
	}
	want := map[string]interface{}{
		"domain": "dom-uuid-5", "product": "DMZ", "bitmask": float64(26),
		"description": "web tier", "network_pool": "pool-uuid-1",
	}
	for k, v := range want {
		if nm.lastPostBody[k] != v {
			t.Errorf("POST body %s = %v, want %v", k, nm.lastPostBody[k], v)
		}
	}
	if _, ok := nm.lastPostBody["prefix"]; ok {
		t.Errorf("POST body should omit prefix, got %v", nm.lastPostBody)
	}
}

func TestAccNetworkResource_ImportAdoptsWithoutReplace(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)
	nm.addNetwork("net-existing", "web tier", "10.20.0.0/26")

	addr := "mcs_network.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config:             networkResourceConfig(mock.URL(), networkDescription),
				ResourceName:       addr,
				ImportState:        true,
				ImportStateId:      "net-existing",
				ImportStatePersist: true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					attrs := states[0].Attributes
					if attrs["domain_uuid"] != "dom-uuid-5" {
						return fmt.Errorf("domain_uuid = %q, want dom-uuid-5", attrs["domain_uuid"])
					}
					if attrs["bitmask"] != "26" {
						return fmt.Errorf("bitmask = %q, want 26", attrs["bitmask"])
					}
					return nil
				},
			},
			{
				// product and network_pool_id are filled in from the configuration in place.
				Config: networkResourceConfig(mock.URL(), networkDescription),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", "net-existing"),
					resource.TestCheckResourceAttr(addr, "product", "DMZ"),
					resource.TestCheckResourceAttr(addr, "network_pool_id", "pool-uuid-1"),
				),
			},
			{
				Config: networkResourceConfig(mock.URL(), networkDescription),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})

	if got := countCalls(mock, http.MethodPost, "/api/v3/networking/networks/"); got != 0 {
		t.Errorf("expected no deploy POST, got %d", got)
	}
}

func TestAccNetworkResource_ChangeForcesReplace(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	newNetworkV3Mock(mock)

	addr := "mcs_network.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: networkResourceConfig(mock.URL(), networkDescription),
				Check:  resource.TestCheckResourceAttr(addr, "id", "net-uuid-1"),
			},
			{
				Config: networkResourceConfig(mock.URL(), `  description = "app tier"`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionDestroyBeforeCreate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", "net-uuid-2"),
					resource.TestCheckResourceAttr(addr, "description", "app tier"),
				),
			},
			{
				// Changing only the timeouts is an in-place, API-free update.
				Config: networkResourceConfig(mock.URL(), `  description = "app tier"
  timeouts {
    create = "45m"
  }`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
				},
				Check: resource.TestCheckResourceAttr(addr, "id", "net-uuid-2"),
			},
		},
	})
}

func TestAccNetworkResource_RemovedOutOfBand(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)

	addr := "mcs_network.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: networkResourceConfig(mock.URL(), networkDescription),
			},
			{
				PreConfig: func() {
					nm.mu.Lock()
					delete(nm.networks, "net-uuid-1")
					nm.mu.Unlock()
				},
				Config: networkResourceConfig(mock.URL(), networkDescription),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionCreate)},
				},
				Check: resource.TestCheckResourceAttr(addr, "id", "net-uuid-2"),
			},
		},
	})
}

func TestAccNetworkResource_DeployFailedTaints(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)
	nm.deployFails = true

	addr := "mcs_network.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config:      networkResourceConfig(mock.URL(), networkDescription),
				ExpectError: regexpMustCompile(`(?s)net-uuid-1 was created but its deploy did not complete.*FortiGate rejected`),
			},
			{
				PreConfig: func() {
					nm.mu.Lock()
					nm.deployFails = false
					nm.mu.Unlock()
				},
				Config: networkResourceConfig(mock.URL(), networkDescription),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionDestroyBeforeCreate)},
				},
				Check: resource.TestCheckResourceAttr(addr, "id", "net-uuid-2"),
			},
		},
	})
}

func TestAccNetworkResource_CreateTimeout(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)
	nm.neverFinish = true

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: networkResourceConfig(mock.URL(), networkDescription+`
  timeouts {
    create = "300ms"
  }`),
				ExpectError: regexpMustCompile(`timed out waiting for network operation: job 101`),
			},
		},
	})
}

func TestAccNetworkResource_DeployPostNotRetried(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)
	nm.postStatus = http.StatusBadGateway

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config:      networkResourceConfig(mock.URL(), networkDescription),
				ExpectError: regexpMustCompile(`(?s)was not retried.*returned status 502`),
			},
		},
	})

	if got := countCalls(mock, http.MethodPost, "/api/v3/networking/networks/"); got != 1 {
		t.Errorf("expected exactly 1 deploy POST, got %d", got)
	}
}

func TestAccNetworkResource_DeleteConflictWaitsForRemoval(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		CheckDestroy: func(_ *terraform.State) error {
			if n := nm.count(); n != 0 {
				return fmt.Errorf("expected the network to be removed, %d left", n)
			}
			return nil
		},
		Steps: []resource.TestStep{
			{
				Config: networkResourceConfig(mock.URL(), networkDescription),
				Check: func(_ *terraform.State) error {
					nm.mu.Lock()
					nm.conflictOnce = true
					nm.mu.Unlock()
					return nil
				},
			},
		},
	})

	if got := countCalls(mock, http.MethodDelete, "/api/v3/networking/networks/"); got != 1 {
		t.Errorf("expected 1 DELETE, got %d", got)
	}
}

func TestAccNetworkResource_Validation(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	cases := map[string]struct {
		config string
		err    string
	}{
		"pool and prefix": {
			config: `
  domain_uuid     = "dom-uuid-5"
  product         = "DMZ"
  bitmask         = 26
  description     = "x"
  network_pool_id = "pool-uuid-1"
  prefix          = "10.0.0.0"`,
			err: `cannot be configured together|Invalid Attribute Combination`,
		},
		"unknown product": {
			config: `
  domain_uuid = "dom-uuid-5"
  product     = "Nope"
  bitmask     = 26
  description = "x"`,
			err: `value must be one of`,
		},
		"bitmask out of range": {
			config: `
  domain_uuid = "dom-uuid-5"
  product     = "DMZ"
  bitmask     = 30
  description = "x"`,
			err: `must be between 24 and 29`,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
				Steps: []resource.TestStep{
					{
						Config:      providerConfigBlock(mock.URL()) + "resource \"mcs_network\" \"test\" {" + tc.config + "\n}",
						PlanOnly:    true,
						ExpectError: regexpMustCompile(tc.err),
					},
				},
			})
		})
	}
}

func TestAccNetworkProductsDataSource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/v3/networking/network-products/", func(w http.ResponseWriter, _ *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `[{"name": "Algemeen"}, {"name": "DMZ"}]`)
	})

	ds := "data.mcs_network_products.all"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_network_products" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ds, "names.#", "2"),
					resource.TestCheckResourceAttr(ds, "names.0", "Algemeen"),
					resource.TestCheckResourceAttr(ds, "names.1", "DMZ"),
				),
			},
		},
	})
}

func TestNetworkOperationState(t *testing.T) {
	str := func(v string) jobNestedString { return jobNestedString{Value: &v} }
	cases := []struct {
		status, result string
		want           networkOperationState
	}{
		{"PROCESSED", "SUCCESS", networkOperationSucceeded},
		{"PROCESSED", "EXCEPTION", networkOperationFailed},
		{"PROCESSED", "SOMETHING_NEW", networkOperationFailed},
		{"RUNNING", "PENDING", networkOperationPending},
		{"QUEUED", "", networkOperationPending},
		{"RUNNING", "EXCEPTION", networkOperationFailed},
		{"", "success", networkOperationSucceeded},
	}
	for _, tc := range cases {
		op := networkOperationAPIModel{Status: str(tc.status), Result: str(tc.result)}
		if got := op.state(); got != tc.want {
			t.Errorf("status %q result %q: got %v, want %v", tc.status, tc.result, got, tc.want)
		}
	}

	// The operations endpoint may return the plain strings of the v3 spec or the nested
	// objects of /api/jobs/job/; both decode the same way.
	for _, body := range []string{
		`{"job_id": 1, "status": "PROCESSED", "result": "EXCEPTION", "message": ""}`,
		`{"job_id": 1, "status": {"status": "PROCESSED"}, "result": {"result": "EXCEPTION"}, "message": ""}`,
	} {
		var op networkOperationAPIModel
		if err := json.Unmarshal([]byte(body), &op); err != nil {
			t.Fatalf("decoding %s: %v", body, err)
		}
		if op.state() != networkOperationFailed {
			t.Errorf("%s: expected failed", body)
		}
		if got := op.describe(); got != `job 1, status "PROCESSED", result "EXCEPTION"` {
			t.Errorf("describe() = %q", got)
		}
	}
}
