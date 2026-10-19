package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func init() {
	networkOperationPollInterval = 10 * time.Millisecond
}

type networkMockJob struct {
	id            int
	undeploy      bool
	dryRun        bool
	deployPlanned bool // a deploy of a planned network via networks/{id}/deploy/
	remaining     int  // polls answered as pending before the job finishes
	networkID     string
}

var networkMockDryRunSteps = []interface{}{
	map[string]interface{}{"name": "apstra_vn", "state": "planned", "attempts": 1, "last_error": "", "changes": []interface{}{
		map[string]interface{}{"system": "apstra", "datacenter": "DC-A", "action": "create", "object": "VN-120", "detail": "virtual network"},
		map[string]interface{}{"system": "apstra", "datacenter": "DC-B", "action": "present", "object": "VN-120", "detail": "already exists"},
	}},
	map[string]interface{}{"name": "fortigate_interface", "state": "planned", "attempts": 1, "last_error": "", "changes": []interface{}{
		map[string]interface{}{"system": "fortigate", "datacenter": nil, "action": "conflict", "object": "vlan120", "detail": "VLAN in use"},
	}},
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
	// plannedDeployFails makes deploys of planned networks fail and roll back.
	plannedDeployFails bool
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
		"deploy_flow": "v1", "state": "active", "leftover_steps": nil,
		"datacenters":          []interface{}{map[string]interface{}{"id": 1, "name": "DC-A"}},
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
	resp := map[string]interface{}{
		"job_id": j.id, "jobname": name, "message": message, "network": network, "dry_run": j.dryRun,
		"status": map[string]interface{}{"status": status}, "result": map[string]interface{}{"result": result},
		"steps": []interface{}{}, "rollback": nil,
	}
	if j.dryRun && result == "SUCCESS" {
		resp["steps"] = networkMockDryRunSteps
	}
	return resp
}

func (m *networkV3Mock) setState(id, state string) {
	if n, ok := m.networks[id]; ok {
		n["state"] = state
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
	if j.deployPlanned {
		if m.plannedDeployFails {
			m.setState(j.networkID, networkStatePlanned)
			resp := m.jobJSON(j, "PROCESSED", "EXCEPTION", "Deploy failed")
			resp["rollback"] = map[string]interface{}{
				"state": "rolled_back", "failed_step": "fortigate_interface", "error": "VLAN in use", "leftover_steps": []string{},
			}
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		m.setState(j.networkID, "active")
		_ = json.NewEncoder(w).Encode(m.jobJSON(j, "PROCESSED", "SUCCESS", "Network deployed"))
		return
	}
	if m.deployFails {
		_ = json.NewEncoder(w).Encode(m.jobJSON(j, "PROCESSED", "EXCEPTION", "FortiGate rejected the interface"))
		return
	}
	if j.dryRun {
		m.setState(j.networkID, networkStatePlanned)
		_ = json.NewEncoder(w).Encode(m.jobJSON(j, "PROCESSED", "SUCCESS", "Dry run finished"))
		return
	}
	m.setState(j.networkID, "active")
	_ = json.NewEncoder(w).Encode(m.jobJSON(j, "PROCESSED", "SUCCESS", "Network deployed"))
}

// handleDeploy serves POST networks/{id}/deploy/ for planned networks.
func (m *networkV3Mock) handleDeploy(w http.ResponseWriter, r *http.Request, id string) {
	n, ok := m.networks[id]
	switch {
	case r.Method != http.MethodPost:
		w.WriteHeader(http.StatusMethodNotAllowed)
	case !ok:
		w.WriteHeader(http.StatusNotFound)
	case n["state"] != networkStatePlanned:
		w.WriteHeader(http.StatusConflict)
		_, _ = fmt.Fprint(w, `{"detail":"The Network is not planned"}`)
	default:
		n["state"] = "deploying"
		m.nextJob++
		j := &networkMockJob{id: m.nextJob, deployPlanned: true, remaining: m.pendingPolls, networkID: id}
		m.jobs[j.id] = j
		w.WriteHeader(http.StatusAccepted)
		_ = json.NewEncoder(w).Encode(m.jobJSON(j, "QUEUED", "PENDING", nil))
	}
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
		n := m.addNetwork(netID, fmt.Sprint(req["description"]), fmt.Sprintf("%s/%v", prefix, req["bitmask"]))
		n["state"] = "deploying"
		if flow, ok := req["deploy_flow"].(string); ok {
			n["deploy_flow"] = flow
		}
		if req["placement"] == "kubernetes" {
			n["datacenters"] = []interface{}{
				map[string]interface{}{"id": 1, "name": "DC-A"}, map[string]interface{}{"id": 2, "name": "DC-K8S"},
			}
		}
		dryRun, _ := req["dry_run"].(bool)

		m.nextJob++
		j := &networkMockJob{id: m.nextJob, dryRun: dryRun, remaining: m.pendingPolls, networkID: netID}
		m.jobs[j.id] = j
		w.WriteHeader(http.StatusAccepted)
		resp := m.jobJSON(j, "QUEUED", "PENDING", nil)
		resp["network"] = nil
		_ = json.NewEncoder(w).Encode(resp)
		return
	}

	if netID, ok := strings.CutSuffix(id, "/deploy"); ok {
		m.handleDeploy(w, r, netID)
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
					resource.TestCheckResourceAttr(addr, "deploy_flow", "v1"),
					resource.TestCheckNoResourceAttr(addr, "placement"),
					resource.TestCheckResourceAttr(addr, "dry_run", "false"),
					resource.TestCheckResourceAttr(addr, "dry_run_changes.#", "0"),
					resource.TestCheckResourceAttr(addr, "datacenters.#", "1"),
					resource.TestCheckResourceAttr(addr, "datacenters.0.id", "1"),
					resource.TestCheckResourceAttr(addr, "datacenters.0.name", "DC-A"),
					resource.TestCheckResourceAttr(addr, "leftover_steps.#", "0"),
					resource.TestCheckNoResourceAttr(addr, "state"),
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
		"description": "web tier", "network_pool": "pool-uuid-1", "deploy_flow": "v1",
	}
	for k, v := range want {
		if nm.lastPostBody[k] != v {
			t.Errorf("POST body %s = %v, want %v", k, nm.lastPostBody[k], v)
		}
	}
	for _, k := range []string{"prefix", "placement", "dry_run"} {
		if _, ok := nm.lastPostBody[k]; ok {
			t.Errorf("POST body should omit %s, got %v", k, nm.lastPostBody)
		}
	}
}

func TestAccNetworkResource_V2KubernetesPlacement(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)

	addr := "mcs_network.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: networkResourceConfig(mock.URL(), networkDescription+`
  deploy_flow     = "v2"
  placement       = "kubernetes"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "deploy_flow", "v2"),
					resource.TestCheckResourceAttr(addr, "placement", "kubernetes"),
					resource.TestCheckResourceAttr(addr, "dry_run", "false"),
					resource.TestCheckResourceAttr(addr, "datacenters.#", "2"),
					resource.TestCheckResourceAttr(addr, "datacenters.1.id", "2"),
					resource.TestCheckResourceAttr(addr, "datacenters.1.name", "DC-K8S"),
				),
			},
			{
				// Import cannot see the placement, so it is filled in from the configuration in place.
				Config:                  networkResourceConfig(mock.URL(), networkDescription+"\n  deploy_flow = \"v2\"\n  placement = \"kubernetes\""),
				ResourceName:            addr,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"product", "network_pool_id", "prefix", "placement"},
			},
			{
				Config: networkResourceConfig(mock.URL(), networkDescription+`
  deploy_flow     = "v2"
  placement       = "standard"`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionDestroyBeforeCreate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", "net-uuid-2"),
					resource.TestCheckResourceAttr(addr, "placement", "standard"),
					resource.TestCheckResourceAttr(addr, "datacenters.#", "1"),
				),
			},
		},
	})

	if nm.lastPostBody["deploy_flow"] != "v2" || nm.lastPostBody["placement"] != "standard" {
		t.Errorf("POST body = %v, want deploy_flow v2 and placement standard", nm.lastPostBody)
	}
	if _, ok := nm.lastPostBody["dry_run"]; ok {
		t.Errorf("POST body should omit dry_run when false, got %v", nm.lastPostBody)
	}
}

func TestAccNetworkResource_V2DefaultPlacement(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)

	addr := "mcs_network.test"
	config := networkResourceConfig(mock.URL(), networkDescription+"\n  deploy_flow = \"v2\"")
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: config,
				Check:  resource.TestCheckResourceAttr(addr, "placement", "standard"),
			},
			{
				// An explicit "standard" matches the resolved default.
				Config: networkResourceConfig(mock.URL(), networkDescription+"\n  deploy_flow = \"v2\"\n  placement = \"standard\""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				Config: config,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
		},
	})

	if nm.lastPostBody["placement"] != "standard" {
		t.Errorf("POST body placement = %v, want standard", nm.lastPostBody["placement"])
	}
}

func TestAccNetworkResource_DryRunThenDeploy(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)

	addr := "mcs_network.test"
	dryRun := func(v bool) string {
		return networkResourceConfig(mock.URL(), networkDescription+fmt.Sprintf(`
  deploy_flow     = "v2"
  dry_run         = %t`, v))
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: dryRun(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", "net-uuid-1"),
					resource.TestCheckResourceAttr(addr, "dry_run", "true"),
					resource.TestCheckResourceAttr(addr, "dry_run_changes.#", "3"),
					resource.TestCheckResourceAttr(addr, "dry_run_changes.0.step", "apstra_vn"),
					resource.TestCheckResourceAttr(addr, "dry_run_changes.0.system", "apstra"),
					resource.TestCheckResourceAttr(addr, "dry_run_changes.0.datacenter", "DC-A"),
					resource.TestCheckResourceAttr(addr, "dry_run_changes.0.action", "create"),
					resource.TestCheckResourceAttr(addr, "dry_run_changes.0.object", "VN-120"),
					resource.TestCheckResourceAttr(addr, "dry_run_changes.0.detail", "virtual network"),
					resource.TestCheckResourceAttr(addr, "dry_run_changes.2.step", "fortigate_interface"),
					resource.TestCheckResourceAttr(addr, "dry_run_changes.2.action", "conflict"),
					resource.TestCheckNoResourceAttr(addr, "dry_run_changes.2.datacenter"),
					func(_ *terraform.State) error {
						nm.mu.Lock()
						defer nm.mu.Unlock()
						if got := nm.networks["net-uuid-1"]["state"]; got != networkStatePlanned {
							return fmt.Errorf("network state = %v, want planned", got)
						}
						return nil
					},
				),
			},
			{
				Config: dryRun(true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
			},
			{
				Config: dryRun(false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", "net-uuid-1"),
					resource.TestCheckResourceAttr(addr, "dry_run", "false"),
					resource.TestCheckResourceAttr(addr, "dry_run_changes.#", "0"),
					func(_ *terraform.State) error {
						nm.mu.Lock()
						defer nm.mu.Unlock()
						if got := nm.networks["net-uuid-1"]["state"]; got != "active" {
							return fmt.Errorf("network state = %v, want active", got)
						}
						return nil
					},
				),
			},
			{
				// A deployed network cannot become planned again.
				Config: dryRun(true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionDestroyBeforeCreate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", "net-uuid-2"),
					resource.TestCheckResourceAttr(addr, "dry_run", "true"),
				),
			},
		},
	})

	if got := countCalls(mock, http.MethodPost, "/api/v3/networking/networks/net-uuid-1/deploy/"); got != 1 {
		t.Errorf("expected 1 deploy of the planned network, got %d", got)
	}
	if nm.lastPostBody["dry_run"] != true {
		t.Errorf("POST body dry_run = %v, want true", nm.lastPostBody["dry_run"])
	}
}

func TestAccNetworkResource_PlannedDeployFails(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)

	addr := "mcs_network.test"
	dryRun := func(v bool) string {
		return networkResourceConfig(mock.URL(), networkDescription+fmt.Sprintf(`
  deploy_flow     = "v2"
  dry_run         = %t`, v))
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: dryRun(true),
			},
			{
				PreConfig: func() {
					nm.mu.Lock()
					nm.plannedDeployFails = true
					nm.mu.Unlock()
				},
				Config: dryRun(false),
				ExpectError: regexpMustCompile(`(?s)Error deploying planned network.*rollback rolled_back after step\s+` +
					`"fortigate_interface" failed: VLAN in use.*still planned`),
			},
			{
				PreConfig: func() {
					nm.mu.Lock()
					nm.plannedDeployFails = false
					nm.mu.Unlock()
				},
				// The failed deploy kept dry_run = true in state, so the deploy is retried in place.
				Config: dryRun(false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", "net-uuid-1"),
					resource.TestCheckResourceAttr(addr, "dry_run", "false"),
				),
			},
		},
	})

	if got := countCalls(mock, http.MethodPost, "/api/v3/networking/networks/net-uuid-1/deploy/"); got != 2 {
		t.Errorf("expected 2 deploys of the planned network, got %d", got)
	}
}

func TestAccNetworkResource_PlannedDeployedOutOfBand(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)

	addr := "mcs_network.test"
	dryRun := func(v bool) string {
		return networkResourceConfig(mock.URL(), networkDescription+fmt.Sprintf(`
  deploy_flow     = "v2"
  dry_run         = %t`, v))
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: dryRun(true),
			},
			{
				PreConfig: func() {
					nm.mu.Lock()
					nm.setState("net-uuid-1", "active")
					nm.mu.Unlock()
				},
				// The deploy endpoint answers 409 because the network is no longer planned.
				Config: dryRun(false),
				Check:  resource.TestCheckResourceAttr(addr, "dry_run", "false"),
			},
		},
	})
}

func TestAccNetworkResource_ImportPlanned(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	nm := newNetworkV3Mock(mock)
	n := nm.addNetwork("net-planned", "web tier", "10.20.0.0/26")
	n["deploy_flow"] = "v2"
	n["state"] = networkStatePlanned

	addr := "mcs_network.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config:             networkResourceConfig(mock.URL(), networkDescription+"\n  deploy_flow = \"v2\"\n  dry_run = true"),
				ResourceName:       addr,
				ImportState:        true,
				ImportStateId:      "net-planned",
				ImportStatePersist: true,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					attrs := states[0].Attributes
					if attrs["dry_run"] != "true" || attrs["deploy_flow"] != "v2" {
						return fmt.Errorf("dry_run = %q, deploy_flow = %q, want true and v2", attrs["dry_run"], attrs["deploy_flow"])
					}
					return nil
				},
			},
			{
				Config: networkResourceConfig(mock.URL(), networkDescription+"\n  deploy_flow = \"v2\"\n  dry_run = true"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
				},
				Check: resource.TestCheckResourceAttr(addr, "placement", "standard"),
			},
		},
	})
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
		"placement with v1": {
			config: `
  domain_uuid = "dom-uuid-5"
  product     = "DMZ"
  bitmask     = 26
  description = "x"
  placement   = "kubernetes"`,
			err: `Placement requires deploy_flow v2`,
		},
		"dry run with explicit v1": {
			config: `
  domain_uuid = "dom-uuid-5"
  product     = "DMZ"
  bitmask     = 26
  description = "x"
  deploy_flow = "v1"
  dry_run     = true`,
			err: `Dry run requires deploy_flow v2`,
		},
		"unknown placement": {
			config: `
  domain_uuid = "dom-uuid-5"
  product     = "DMZ"
  bitmask     = 26
  description = "x"
  deploy_flow = "v2"
  placement   = "edge"`,
			err: `value must be one of`,
		},
		"unknown deploy flow": {
			config: `
  domain_uuid = "dom-uuid-5"
  product     = "DMZ"
  bitmask     = 26
  description = "x"
  deploy_flow = "v3"`,
			err: `value must be one of`,
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

	var op networkOperationAPIModel
	body := `{"job_id": 7, "jobname": "deploy_network", "status": "PROCESSED", "result": "EXCEPTION", "message": "Deploy failed",
		"dry_run": false, "steps": [{"name": "apstra_vn", "state": "done", "attempts": 2, "last_error": "", "changes": []}],
		"rollback": {"state": "rollback_failed", "failed_step": "fortigate_interface", "error": "VLAN in use",
			"leftover_steps": ["apstra_vn", "infoblox"]}}`
	if err := json.Unmarshal([]byte(body), &op); err != nil {
		t.Fatalf("decoding rollback: %v", err)
	}
	want := `job 7 (deploy_network), status "PROCESSED", result "EXCEPTION": Deploy failed; rollback rollback_failed ` +
		`after step "fortigate_interface" failed: VLAN in use (configuration may remain for: apstra_vn, infoblox)`
	if got := op.describe(); got != want {
		t.Errorf("describe() = %q, want %q", got, want)
	}
	if len(op.Steps) != 1 || op.Steps[0].Attempts != 2 {
		t.Errorf("steps = %+v", op.Steps)
	}
}

func TestLeftoverStepsList(t *testing.T) {
	cases := map[string][]string{
		`null`:                       {},
		``:                           {},
		`["apstra_vn", "infoblox"]`:  {"apstra_vn", "infoblox"},
		`[{"name": "apstra_vn"}]`:    {"apstra_vn"},
		`[1, {"step": "fortigate"}]`: {"1", `{"step":"fortigate"}`},
		`"not a list"`:               {},
	}
	for raw, want := range cases {
		got := leftoverStepsList(json.RawMessage(raw)).Elements()
		if len(got) != len(want) {
			t.Errorf("%s: got %v, want %v", raw, got, want)
			continue
		}
		for i, v := range got {
			if s := v.(types.String).ValueString(); s != want[i] {
				t.Errorf("%s: element %d = %q, want %q", raw, i, s, want[i])
			}
		}
	}
}

func TestAccNetworkPlacementsDataSource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/v3/networking/network-placements/", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("domain") != "dom-uuid-5" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"detail":"Domain not found"}`)
			return
		}
		_, _ = fmt.Fprint(w, `{"standard": [{"id": 1, "name": "DC-A"}],
			"kubernetes": [{"id": 1, "name": "DC-A"}, {"id": 2, "name": "DC-K8S"}],
			"kubernetes_equals_standard": false}`)
	})

	ds := "data.mcs_network_placements.prod"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_network_placements" "prod" {
  domain_uuid = "dom-uuid-5"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(ds, "domain_uuid", "dom-uuid-5"),
					resource.TestCheckResourceAttr(ds, "standard.#", "1"),
					resource.TestCheckResourceAttr(ds, "standard.0.id", "1"),
					resource.TestCheckResourceAttr(ds, "standard.0.name", "DC-A"),
					resource.TestCheckResourceAttr(ds, "kubernetes.#", "2"),
					resource.TestCheckResourceAttr(ds, "kubernetes.1.id", "2"),
					resource.TestCheckResourceAttr(ds, "kubernetes.1.name", "DC-K8S"),
					resource.TestCheckResourceAttr(ds, "kubernetes_equals_standard", "false"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_network_placements" "prod" {
  domain_uuid = "dom-uuid-missing"
}`,
				ExpectError: regexpMustCompile(`(?s)Error reading network placements.*404`),
			},
		},
	})
}
