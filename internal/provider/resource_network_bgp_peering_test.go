package provider

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const (
	bgpTestNetwork = "net-uuid-8"
	bgpJobname     = "change_network_bgp_v2"
)

type bgpMockJob struct {
	id        int
	remaining int    // polls answered as pending before the job finishes
	fail      string // when set, the job fails and is rolled back with this error
	apply     func() // what the finished job changes
	done      bool
}

// bgpV3Mock serves the v3 BGP peering endpoint of one network. Every write answers 202 with an
// operation that stays pending for pendingPolls polls before it is applied.
type bgpV3Mock struct {
	mu           sync.Mutex
	enabled      bool
	prefixes     []string
	clusters     []string // Kubernetes clusters using the network: DELETE answers 409 while set
	putConflict  string   // when set, PUT answers 409 with this detail
	failNext     string
	jobs         map[int]*bgpMockJob
	nextJob      int
	pendingPolls int
	puts         []string // request bodies of the PUTs
	deletes      int
}

func newBgpV3Mock(mock *mockAPIServer) *bgpV3Mock {
	m := &bgpV3Mock{jobs: map[int]*bgpMockJob{}, nextJob: 700, pendingPolls: 2, prefixes: []string{}}
	mock.On("/api/v3/networking/operations/", m.handleOperation)
	mock.On("/api/v3/networking/networks/", m.handlePeering)
	return m
}

func (m *bgpV3Mock) queue(apply func()) *bgpMockJob {
	m.nextJob++
	j := &bgpMockJob{id: m.nextJob, remaining: m.pendingPolls, apply: apply, fail: m.failNext}
	m.failNext = ""
	m.jobs[j.id] = j
	return j
}

func (m *bgpV3Mock) jobJSON(j *bgpMockJob) map[string]interface{} {
	status, result := "QUEUED", "PENDING"
	var rollback interface{}
	switch {
	case j.done && j.fail != "":
		status, result = "PROCESSED", "EXCEPTION"
		rollback = map[string]interface{}{
			"state": "rolled_back", "failed_step": "fortigate_bgp", "error": j.fail, "leftover_steps": []string{},
		}
	case j.done:
		status, result = "PROCESSED", "SUCCESS"
	}
	return map[string]interface{}{
		"job_id": j.id, "jobname": bgpJobname, "dry_run": false, "message": nil, "network": nil,
		"status": map[string]interface{}{"status": status}, "result": map[string]interface{}{"result": result},
		"dhcp_lease": nil, "steps": []interface{}{}, "rollback": rollback, "warnings": []string{},
	}
}

func (m *bgpV3Mock) handleOperation(w http.ResponseWriter, r *http.Request, _ []byte) {
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

func (m *bgpV3Mock) peeringJSON() map[string]interface{} {
	if !m.enabled {
		return map[string]interface{}{"enabled": false, "neighbor_group": nil, "accepted_prefixes": []string{}}
	}
	return map[string]interface{}{"enabled": true, "neighbor_group": "MCS-K8S-120", "accepted_prefixes": m.prefixes}
}

func (m *bgpV3Mock) handlePeering(w http.ResponseWriter, r *http.Request, body []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	netID, ok := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/api/v3/networking/networks/"), "/bgp-peering/")
	if !ok || netID != bgpTestNetwork {
		writeJSON(w, http.StatusNotFound, map[string]string{"detail": "Not found."})
		return
	}
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, http.StatusOK, m.peeringJSON())
	case http.MethodPut:
		m.puts = append(m.puts, string(body))
		if m.putConflict != "" {
			writeJSON(w, http.StatusConflict, map[string]interface{}{"detail": m.putConflict, "state": "deploying"})
			return
		}
		j := m.queue(func() { m.enabled = true })
		writeJSON(w, http.StatusAccepted, m.jobJSON(j))
	case http.MethodDelete:
		m.deletes++
		if len(m.clusters) > 0 {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"detail":   fmt.Sprintf("Kubernetes Cluster %s uses this Network; switch BGP peering off once that cluster is destroyed.", strings.Join(m.clusters, ", ")),
				"clusters": m.clusters,
			})
			return
		}
		j := m.queue(func() { m.enabled = false })
		writeJSON(w, http.StatusAccepted, m.jobJSON(j))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (m *bgpV3Mock) set(f func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	f()
}

func (m *bgpV3Mock) checkOff(_ *terraform.State) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.enabled {
		return fmt.Errorf("expected BGP peering to be switched off")
	}
	return nil
}

func bgpPeeringConfig(url, extra string) string {
	return providerConfigBlock(url) + fmt.Sprintf(`
resource "mcs_network_bgp_peering" "test" {
  network = %q
%s
}
`, bgpTestNetwork, extra)
}

func TestNetworkBgpPeering_Lifecycle(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	bm := newBgpV3Mock(mock)

	addr := "mcs_network_bgp_peering.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		CheckDestroy:             bm.checkOff,
		Steps: []resource.TestStep{
			{
				Config: bgpPeeringConfig(mock.URL(), ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "id", bgpTestNetwork),
					resource.TestCheckResourceAttr(addr, "network", bgpTestNetwork),
					resource.TestCheckResourceAttr(addr, "neighbor_group", "MCS-K8S-120"),
					resource.TestCheckResourceAttr(addr, "accepted_prefixes.#", "0"),
				),
			},
			{
				// A cluster on the network adds its load balancer pool; the refresh picks it up
				// without planning a change.
				PreConfig: func() { bm.set(func() { bm.prefixes = []string{"10.50.0.0/28"} }) },
				Config:    bgpPeeringConfig(mock.URL(), ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "accepted_prefixes.#", "1"),
					resource.TestCheckResourceAttr(addr, "accepted_prefixes.0", "10.50.0.0/28"),
				),
			},
			{
				// Changing the timeouts switches peering on again and waits for it.
				Config: bgpPeeringConfig(mock.URL(), `  timeouts {
    update = "30m"
  }`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionUpdate)},
				},
				Check: resource.TestCheckResourceAttr(addr, "neighbor_group", "MCS-K8S-120"),
			},
			{
				Config:                  bgpPeeringConfig(mock.URL(), ""),
				ResourceName:            addr,
				ImportState:             true,
				ImportStateId:           bgpTestNetwork,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"timeouts"},
			},
		},
	})

	if len(bm.puts) != 2 {
		t.Fatalf("expected 2 PUTs (create and update), got %d", len(bm.puts))
	}
	for _, b := range bm.puts {
		if b != "" {
			t.Errorf("PUT must not send a body, got %q", b)
		}
	}
	if bm.deletes != 1 {
		t.Errorf("expected 1 DELETE on destroy, got %d", bm.deletes)
	}
}

func TestNetworkBgpPeering_SwitchedOffOutOfBandIsRecreated(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	bm := newBgpV3Mock(mock)

	addr := "mcs_network_bgp_peering.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{Config: bgpPeeringConfig(mock.URL(), "")},
			{
				PreConfig: func() { bm.set(func() { bm.enabled = false }) },
				Config:    bgpPeeringConfig(mock.URL(), ""),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction(addr, plancheck.ResourceActionCreate)},
				},
				Check: resource.TestCheckResourceAttr(addr, "neighbor_group", "MCS-K8S-120"),
			},
		},
	})
	if len(bm.puts) != 2 {
		t.Errorf("expected 2 PUTs, got %d", len(bm.puts))
	}
}

func TestNetworkBgpPeering_DestroyRefusedWhileClustersUseTheNetwork(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	bm := newBgpV3Mock(mock)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		CheckDestroy:             bm.checkOff,
		Steps: []resource.TestStep{
			{
				Config: bgpPeeringConfig(mock.URL(), ""),
			},
			{
				// Removing the resource is refused with MCS's own reason, naming the cluster.
				PreConfig: func() { bm.set(func() { bm.clusters = []string{"k8s-prod"} }) },
				Config:    providerConfigBlock(mock.URL()),
				ExpectError: regexp.MustCompile(`(?s)Cannot switch off network BGP peering.*` +
					`Kubernetes\s+Cluster\s+k8s-prod\s+uses\s+this\s+Network`),
			},
			{
				// Once the cluster is gone, the same apply switches peering off.
				PreConfig: func() { bm.set(func() { bm.clusters = nil }) },
				Config:    providerConfigBlock(mock.URL()),
			},
		},
	})

	if bm.deletes != 2 {
		t.Errorf("expected 2 DELETEs (refused, then accepted), got %d", bm.deletes)
	}
}

func TestNetworkBgpPeering_PutConflictIsReadable(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	bm := newBgpV3Mock(mock)
	bm.putConflict = "A deploy or teardown of this Network is queued or running."

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config: bgpPeeringConfig(mock.URL(), ""),
			ExpectError: regexp.MustCompile(`(?s)MCS\s+refused\s+to\s+switch\s+BGP\s+peering\s+on\s+for\s+network\s+net-uuid-8:\s+` +
				`A\s+deploy\s+or\s+teardown\s+of\s+this\s+Network\s+is\s+queued\s+or\s+running`),
		}},
	})
	if bm.enabled {
		t.Errorf("a refused PUT must not switch peering on")
	}
}

func TestNetworkBgpPeering_FailedChangeReportsTheRollback(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	bm := newBgpV3Mock(mock)
	bm.failNext = "FortiGate refused the neighbor-group"

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config: bgpPeeringConfig(mock.URL(), ""),
			ExpectError: regexp.MustCompile(`(?s)rollback rolled_back after step\s+"fortigate_bgp" failed:\s+FortiGate\s+` +
				`refused\s+the\s+neighbor-group.*The change was rolled back`),
		}},
	})
	if bm.enabled {
		t.Errorf("a failed change must not switch peering on")
	}
}
