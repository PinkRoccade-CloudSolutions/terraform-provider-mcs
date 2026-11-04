package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	dhcpTestNetwork = "net-uuid-7"
	dhcpJobname     = "change_network_dhcp_v2"
)

type dhcpMockJob struct {
	id        int
	remaining int    // polls answered as pending before the job finishes
	fail      string // when set, the job fails and is rolled back with this error
	apply     func() // what the finished job changes
	leaseID   string // lease the job writes, carried by the operation once the row exists
	done      bool
}

// dhcpV3Mock serves the v3 DHCP server and static lease endpoints. Every write answers 202 with
// an operation that stays pending for pendingPolls polls before it is applied.
type dhcpV3Mock struct {
	mu           sync.Mutex
	servers      map[string]map[string]interface{} // enabled servers by network id
	leases       map[string]map[string]interface{}
	jobs         map[int]*dhcpMockJob
	nextJob      int
	nextLease    int
	pendingPolls int
	pageSize     int
	failNext     string
	serverPuts   []map[string]interface{}
	leasePosts   []map[string]interface{}
	leasePatches []map[string]interface{}
	listQueries  []string
}

func newDhcpV3Mock(mock *mockAPIServer) *dhcpV3Mock {
	m := &dhcpV3Mock{
		servers:      map[string]map[string]interface{}{},
		leases:       map[string]map[string]interface{}{},
		jobs:         map[int]*dhcpMockJob{},
		nextJob:      500,
		pendingPolls: 2,
		pageSize:     2,
	}
	mock.On("/api/v3/networking/operations/", m.handleOperation)
	mock.On("/api/v3/networking/networks/", m.handleServer)
	mock.On(dhcpLeasesPath, m.handleLeases)
	return m
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func (m *dhcpV3Mock) queue(apply func()) *dhcpMockJob {
	m.nextJob++
	j := &dhcpMockJob{id: m.nextJob, remaining: m.pendingPolls, apply: apply, fail: m.failNext}
	m.failNext = ""
	m.jobs[j.id] = j
	return j
}

func (m *dhcpV3Mock) jobJSON(j *dhcpMockJob) map[string]interface{} {
	status, result := "QUEUED", "PENDING"
	var rollback interface{}
	switch {
	case j.done && j.fail != "":
		status, result = "PROCESSED", "EXCEPTION"
		rollback = map[string]interface{}{
			"state": "rolled_back", "failed_step": "fortigate_dhcp", "error": j.fail, "leftover_steps": []string{},
		}
	case j.done:
		status, result = "PROCESSED", "SUCCESS"
	}
	var lease interface{}
	if l, ok := m.leases[j.leaseID]; ok && j.done {
		lease = l
	}
	return map[string]interface{}{
		"job_id": j.id, "jobname": dhcpJobname, "dry_run": false, "message": nil, "network": nil,
		"status": map[string]interface{}{"status": status}, "result": map[string]interface{}{"result": result},
		"dhcp_lease": lease, "steps": []interface{}{}, "rollback": rollback, "warnings": []string{},
	}
}

func (m *dhcpV3Mock) handleOperation(w http.ResponseWriter, r *http.Request, _ []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	var id int
	_, _ = fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/api/v3/networking/operations/"), "%d", &id)
	j, ok := m.jobs[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"detail": "Not found."})
		return
	}
	if j.remaining > 0 {
		j.remaining--
	} else if !j.done {
		j.done = true
		if j.fail == "" {
			j.apply()
		}
	}
	writeJSON(w, http.StatusOK, m.jobJSON(j))
}

func (m *dhcpV3Mock) serverJSON(netID string) map[string]interface{} {
	if s, ok := m.servers[netID]; ok {
		out := map[string]interface{}{"network": netID, "enabled": true, "netmask": "255.255.255.0", "default_gateway": "10.20.30.1"}
		for k, v := range s {
			out[k] = v
		}
		return out
	}
	return map[string]interface{}{
		"network": netID, "enabled": false, "pool_start": nil, "pool_end": nil,
		"netmask": "255.255.255.0", "default_gateway": "10.20.30.1", "dns_servers": []string{}, "lease_time": nil,
	}
}

func (m *dhcpV3Mock) handleServer(w http.ResponseWriter, r *http.Request, body []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	netID, ok := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/api/v3/networking/networks/"), "/dhcp-server/")
	if !ok || netID != dhcpTestNetwork {
		writeJSON(w, http.StatusNotFound, map[string]string{"detail": "Not found."})
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, m.serverJSON(netID))
	case http.MethodPut:
		var req map[string]interface{}
		_ = json.Unmarshal(body, &req)
		m.serverPuts = append(m.serverPuts, req)
		j := m.queue(func() { m.servers[netID] = req })
		writeJSON(w, http.StatusAccepted, m.jobJSON(j))
	case http.MethodDelete:
		for _, l := range m.leases {
			if l["network"] == netID {
				writeJSON(w, http.StatusConflict, map[string]string{"detail": "The DHCP server still has static leases."})
				return
			}
		}
		j := m.queue(func() { delete(m.servers, netID) })
		writeJSON(w, http.StatusAccepted, m.jobJSON(j))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *dhcpV3Mock) addLease(id, netID, ip, mac string) map[string]interface{} {
	l := map[string]interface{}{
		"id": id, "network": netID, "ip": ip, "mac": normalizeMAC(mac), "description": "",
		"created_at_timestamp": "2026-11-04T10:00:00Z", "updated_at_timestamp": "2026-11-04T10:00:00Z",
	}
	m.leases[id] = l
	return l
}

func (m *dhcpV3Mock) handleLeases(w http.ResponseWriter, r *http.Request, body []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	id := strings.Trim(strings.TrimPrefix(r.URL.Path, dhcpLeasesPath), "/")
	var req map[string]interface{}
	_ = json.Unmarshal(body, &req)

	if id == "" {
		switch r.Method {
		case http.MethodGet:
			m.listLeases(w, r)
		case http.MethodPost:
			m.leasePosts = append(m.leasePosts, req)
			netID, _ := req["network"].(string)
			if _, ok := m.servers[netID]; !ok {
				writeJSON(w, http.StatusBadRequest, map[string][]string{"network": {"DHCP is not enabled on this Network."}})
				return
			}
			m.nextLease++
			leaseID := fmt.Sprintf("lease-uuid-%d", m.nextLease)
			j := m.queue(nil)
			j.leaseID = leaseID
			j.apply = func() {
				l := m.addLease(leaseID, netID, fmt.Sprint(req["ip"]), fmt.Sprint(req["mac"]))
				l["description"] = req["description"]
			}
			writeJSON(w, http.StatusAccepted, m.jobJSON(j))
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
		return
	}

	l, ok := m.leases[id]
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"detail": "Not found."})
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, l)
	case http.MethodPatch:
		m.leasePatches = append(m.leasePatches, req)
		j := m.queue(func() {
			for k, v := range req {
				if k == "mac" {
					v = normalizeMAC(fmt.Sprint(v))
				}
				l[k] = v
			}
			l["updated_at_timestamp"] = "2026-11-04T11:00:00Z"
		})
		j.leaseID = id
		writeJSON(w, http.StatusAccepted, m.jobJSON(j))
	case http.MethodDelete:
		j := m.queue(func() { delete(m.leases, id) })
		writeJSON(w, http.StatusAccepted, m.jobJSON(j))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// listLeases answers the paginated, filterable lease list with absolute `next` links.
func (m *dhcpV3Mock) listLeases(w http.ResponseWriter, r *http.Request) {
	m.listQueries = append(m.listQueries, r.URL.RawQuery)
	q := r.URL.Query()
	var match []map[string]interface{}
	for _, l := range m.leases {
		if (q.Get("network") == "" || l["network"] == q.Get("network")) &&
			(q.Get("ip") == "" || l["ip"] == q.Get("ip")) && (q.Get("mac") == "" || l["mac"] == q.Get("mac")) {
			match = append(match, l)
		}
	}
	sort.Slice(match, func(i, k int) bool { return fmt.Sprint(match[i]["id"]) < fmt.Sprint(match[k]["id"]) })
	offset := 0
	_, _ = fmt.Sscanf(q.Get("offset"), "%d", &offset)
	end := min(offset+m.pageSize, len(match))
	var next interface{}
	if end < len(match) {
		q.Set("offset", fmt.Sprint(end))
		next = "http://" + r.Host + r.URL.Path + "?" + q.Encode()
	}
	page := match[min(offset, len(match)):end]
	if page == nil {
		page = []map[string]interface{}{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"count": len(match), "next": next, "previous": nil, "results": page})
}

func (m *dhcpV3Mock) checkAllRemoved(_ *terraform.State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if len(m.servers) != 0 || len(m.leases) != 0 {
		return fmt.Errorf("expected DHCP server and leases to be removed, have %d servers and %d leases", len(m.servers), len(m.leases))
	}
	return nil
}

func dhcpServerConfig(url, leaseTime, dns string) string {
	return providerConfigBlock(url) + fmt.Sprintf(`
resource "mcs_network_dhcp_server" "test" {
  network     = %q
  pool_start  = "10.20.30.100"
  pool_end    = "10.20.30.200"
  dns_servers = [%s]
  lease_time  = %s
}
`, dhcpTestNetwork, dns, leaseTime)
}

func TestNetworkDhcpServer_Lifecycle(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	dm := newDhcpV3Mock(mock)

	addr := "mcs_network_dhcp_server.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		CheckDestroy:             dm.checkAllRemoved,
		Steps: []resource.TestStep{
			{
				Config: dhcpServerConfig(mock.URL(), "86400", `"10.1.1.1"`),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", dhcpTestNetwork),
					resource.TestCheckResourceAttr(addr, "network", dhcpTestNetwork),
					resource.TestCheckResourceAttr(addr, "pool_start", "10.20.30.100"),
					resource.TestCheckResourceAttr(addr, "pool_end", "10.20.30.200"),
					resource.TestCheckResourceAttr(addr, "dns_servers.#", "1"),
					resource.TestCheckResourceAttr(addr, "lease_time", "86400"),
					resource.TestCheckResourceAttr(addr, "netmask", "255.255.255.0"),
					resource.TestCheckResourceAttr(addr, "default_gateway", "10.20.30.1"),
				),
			},
			{
				Config: dhcpServerConfig(mock.URL(), "3600", `"10.1.1.1", "10.1.1.2"`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "lease_time", "3600"),
					resource.TestCheckResourceAttr(addr, "dns_servers.1", "10.1.1.2"),
				),
			},
			{
				Config:            dhcpServerConfig(mock.URL(), "3600", `"10.1.1.1", "10.1.1.2"`),
				ResourceName:      addr,
				ImportState:       true,
				ImportStateId:     dhcpTestNetwork,
				ImportStateVerify: true,
			},
		},
	})

	if len(dm.serverPuts) != 2 {
		t.Fatalf("expected 2 server PUTs, got %d", len(dm.serverPuts))
	}
	got, _ := json.Marshal(dm.serverPuts[1])
	want := `{"dns_servers":["10.1.1.1","10.1.1.2"],"lease_time":3600,"pool_end":"10.20.30.200","pool_start":"10.20.30.100"}`
	if string(got) != want {
		t.Errorf("update PUT body = %s, want %s", got, want)
	}
}

func TestNetworkDhcpServer_RefusesToTakeOverAnExistingServer(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	dm := newDhcpV3Mock(mock)
	dm.servers[dhcpTestNetwork] = map[string]interface{}{
		"pool_start": "10.20.30.10", "pool_end": "10.20.30.20", "dns_servers": []string{"10.1.1.1"}, "lease_time": 600,
	}

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config:      dhcpServerConfig(mock.URL(), "86400", `"10.1.1.1"`),
			ExpectError: regexp.MustCompile(`already runs a DHCP server`),
		}},
	})
	if len(dm.serverPuts) != 0 {
		t.Errorf("expected no PUT, got %v", dm.serverPuts)
	}
}

func TestNetworkDhcpServer_FailedChangeReportsTheRollback(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	dm := newDhcpV3Mock(mock)
	dm.failNext = "FortiGate refused the ip-range"

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config: dhcpServerConfig(mock.URL(), "86400", `"10.1.1.1"`),
			ExpectError: regexp.MustCompile(`(?s)rollback rolled_back after step\s+"fortigate_dhcp" failed:\s+FortiGate\s+` +
				`refused\s+the\s+ip-range.*The change was rolled back`),
		}},
	})
	if len(dm.servers) != 0 {
		t.Errorf("a failed change must not enable the server")
	}
}

func dhcpServerAndLeaseConfig(url, mac, description string) string {
	return dhcpServerConfig(url, "86400", `"10.1.1.1"`) + fmt.Sprintf(`
resource "mcs_network_dhcp_lease" "printer" {
  network     = mcs_network_dhcp_server.test.network
  ip          = "10.20.30.11"
  mac         = %q
  description = %q
}
`, mac, description)
}

func TestNetworkDhcpLease_LifecycleWithServer(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	dm := newDhcpV3Mock(mock)

	addr := "mcs_network_dhcp_lease.printer"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		// The server answers 409 to a DELETE while leases exist, so destroy only passes when the
		// lease goes first.
		CheckDestroy: dm.checkAllRemoved,
		Steps: []resource.TestStep{
			{
				// The MAC is stored as 00:50:56:aa:bb:01; the configured notation must not show a diff.
				Config: dhcpServerAndLeaseConfig(mock.URL(), "00-50-56-AA-BB-01", "printer"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", "lease-uuid-1"),
					resource.TestCheckResourceAttr(addr, "network", dhcpTestNetwork),
					resource.TestCheckResourceAttr(addr, "ip", "10.20.30.11"),
					resource.TestCheckResourceAttr(addr, "mac", "00-50-56-AA-BB-01"),
					resource.TestCheckResourceAttr(addr, "description", "printer"),
					resource.TestCheckResourceAttr(addr, "created_at", "2026-11-04T10:00:00Z"),
				),
			},
			{
				// Rewriting the MAC in another notation updates state only; nothing is sent to MCS.
				Config: dhcpServerAndLeaseConfig(mock.URL(), "0050.56aa.bb01", "printer"),
				Check:  resource.TestCheckResourceAttr(addr, "mac", "0050.56aa.bb01"),
			},
			{
				Config: dhcpServerAndLeaseConfig(mock.URL(), "00-50-56-AA-BB-01", "printer 2nd floor"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "description", "printer 2nd floor"),
					resource.TestCheckResourceAttr(addr, "updated_at", "2026-11-04T11:00:00Z"),
				),
			},
			{
				Config:            dhcpServerAndLeaseConfig(mock.URL(), "00-50-56-AA-BB-01", "printer 2nd floor"),
				ResourceName:      addr,
				ImportState:       true,
				ImportStateVerify: true,
				// Import reads the stored notation.
				ImportStateVerifyIgnore: []string{"mac"},
			},
		},
	})

	if len(dm.leasePosts) != 1 {
		t.Fatalf("expected 1 lease POST, got %d", len(dm.leasePosts))
	}
	if got := dm.leasePosts[0]["network"]; got != dhcpTestNetwork {
		t.Errorf("POST network = %v", got)
	}
	if len(dm.leasePatches) != 1 {
		t.Fatalf("expected 1 lease PATCH (the description change), got %d: %v", len(dm.leasePatches), dm.leasePatches)
	}
	if got := dm.leasePatches[0]["description"]; got != "printer 2nd floor" {
		t.Errorf("PATCH description = %v", got)
	}
	if _, ok := dm.leasePatches[0]["network"]; ok {
		t.Errorf("PATCH must not send network: %v", dm.leasePatches[0])
	}
	if got := countCalls(mock, http.MethodPut, dhcpLeasesPath); got != 0 {
		t.Errorf("leases must be changed with PATCH, got %d PUTs", got)
	}
}

func TestNetworkDhcpLease_ChangingNetworkReplaces(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	dm := newDhcpV3Mock(mock)
	dm.servers[dhcpTestNetwork] = map[string]interface{}{}
	dm.servers["net-uuid-other"] = map[string]interface{}{}
	dm.addLease("lease-uuid-9", dhcpTestNetwork, "10.20.30.12", "00:50:56:aa:bb:02")

	cfg := func(network string) string {
		return providerConfigBlock(mock.URL()) + fmt.Sprintf(`
resource "mcs_network_dhcp_lease" "x" {
  network = %q
  ip      = "10.20.30.12"
  mac     = "00:50:56:aa:bb:02"
}
`, network)
	}
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config:             cfg(dhcpTestNetwork),
				ResourceName:       "mcs_network_dhcp_lease.x",
				ImportState:        true,
				ImportStateId:      "lease-uuid-9",
				ImportStatePersist: true,
			},
			{
				Config: cfg("net-uuid-other"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("mcs_network_dhcp_lease.x", plancheck.ResourceActionDestroyBeforeCreate)},
				},
				Check: resource.TestCheckResourceAttr("mcs_network_dhcp_lease.x", "network", "net-uuid-other"),
			},
		},
	})
}

func TestNetworkDhcpLeasesDataSource_FollowsPagesAndFilters(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	dm := newDhcpV3Mock(mock)
	dm.addLease("lease-a", dhcpTestNetwork, "10.20.30.11", "00:50:56:aa:bb:01")
	dm.addLease("lease-b", dhcpTestNetwork, "10.20.30.12", "00:50:56:aa:bb:02")
	dm.addLease("lease-c", dhcpTestNetwork, "10.20.30.13", "00:50:56:aa:bb:03")
	dm.addLease("lease-d", "net-uuid-other", "10.9.9.9", "00:50:56:aa:bb:04")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config: providerConfigBlock(mock.URL()) + fmt.Sprintf(`
data "mcs_network_dhcp_leases" "all" {
  network = %[1]q
}

data "mcs_network_dhcp_leases" "by_mac" {
  network = %[1]q
  mac     = "00-50-56-AA-BB-02"
}
`, dhcpTestNetwork),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.mcs_network_dhcp_leases.all", "leases.#", "3"),
				resource.TestCheckResourceAttr("data.mcs_network_dhcp_leases.all", "leases.2.id", "lease-c"),
				resource.TestCheckResourceAttr("data.mcs_network_dhcp_leases.by_mac", "leases.#", "1"),
				resource.TestCheckResourceAttr("data.mcs_network_dhcp_leases.by_mac", "leases.0.ip", "10.20.30.12"),
			),
		}},
	})

	for _, q := range dm.listQueries {
		if strings.Contains(q, "mac=") && !strings.Contains(q, "mac=00%3A50%3A56%3Aaa%3Abb%3A02") {
			t.Errorf("mac filter should be sent normalised, got %q", q)
		}
	}
}

func TestNormalizeMAC(t *testing.T) {
	for in, want := range map[string]string{
		"00:50:56:aa:bb:01": "00:50:56:aa:bb:01",
		"00-50-56-AA-BB-01": "00:50:56:aa:bb:01",
		"0050.56aa.bb01":    "00:50:56:aa:bb:01",
		"00:50:56:aa:bb":    "",
		"00:00:00:00:fe:80:00:00:00:00:00:00:02:00:5e:10:00:00:00:01": "",
		"printer": "",
	} {
		if got := normalizeMAC(in); got != want {
			t.Errorf("normalizeMAC(%q) = %q, want %q", in, got, want)
		}
	}
}
