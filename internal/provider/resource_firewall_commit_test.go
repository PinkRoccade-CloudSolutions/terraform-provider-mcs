package provider

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

const commitDomainUUID = "6f1c2c1e-7c3b-4b8e-9d1e-0a1b2c3d4e5f"

// commitDomainMock serves the domain lookup with one domain named VDOM-001 of the given flow.
func commitDomainMock(m *mockAPIServer, flow string) {
	m.On("/api/tenant/domains/", func(w http.ResponseWriter, _ *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprintf(w, `{"count":2,"next":null,"results":[
			{"id":2,"uuid":"other","name":"VDOM-0011","deploy_flow":"v1","state":"active"},
			{"id":1,"uuid":%q,"name":"VDOM-001","deploy_flow":%q,"state":"active"}]}`, commitDomainUUID, flow)
	})
}

func checkCallCount(m *mockAPIServer, method, path string, want int) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		if got := countCalls(m, method, path); got != want {
			return fmt.Errorf("expected %d %s %s calls, got %d: %v", want, method, path, got, m.Calls())
		}
		return nil
	}
}

func commitConfig(url, changeID, trigger string) string {
	return providerConfigBlock(url) + fmt.Sprintf(`
resource "mcs_firewall_commit" "test" {
  domain   = "VDOM-001"
  changeid = %q
  triggers = { rules = %q }
}`, changeID, trigger)
}

const v1CommitPath = "/api/networking/domain/VDOM-001/commit/"

func TestAccFirewallCommitResource_V1(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	commitDomainMock(mock, "v1")
	mock.On(v1CommitPath, func(w http.ResponseWriter, _ *http.Request, body []byte) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: commitConfig(mock.URL(), "CHG-1", "a"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_commit.test", "deploy_flow", "v1"),
					resource.TestCheckResourceAttr("mcs_firewall_commit.test", "changeid", "CHG-1"),
					resource.TestCheckNoResourceAttr("mcs_firewall_commit.test", "job_id"),
					resource.TestCheckResourceAttrSet("mcs_firewall_commit.test", "committed_at"),
					resource.TestMatchResourceAttr("mcs_firewall_commit.test", "id", regexp.MustCompile(`^VDOM-001/`)),
					fwCheckCall(mock, http.MethodPost, v1CommitPath, map[string]string{"changeid": `"CHG-1"`}),
					checkCallCount(mock, http.MethodPost, v1CommitPath, 1),
				),
			},
			{
				// Only the change ID changes: stored in place, no new commit.
				Config: commitConfig(mock.URL(), "CHG-2", "a"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("mcs_firewall_commit.test", plancheck.ResourceActionUpdate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_firewall_commit.test", "changeid", "CHG-2"),
					checkCallCount(mock, http.MethodPost, v1CommitPath, 1),
				),
			},
			{
				// A trigger changes: the resource is replaced and commits again.
				Config: commitConfig(mock.URL(), "CHG-2", "b"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{plancheck.ExpectResourceAction("mcs_firewall_commit.test", plancheck.ResourceActionDestroyBeforeCreate)},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkCallCount(mock, http.MethodPost, v1CommitPath, 2),
				),
			},
		},
		CheckDestroy: func(_ *terraform.State) error {
			for _, c := range mock.Calls() {
				if c.Method == http.MethodDelete {
					return fmt.Errorf("destroy should not call the API, got %s %s", c.Method, c.Path)
				}
			}
			return checkCallCount(mock, http.MethodPost, v1CommitPath, 2)(nil)
		},
	})
}

func TestAccFirewallCommitResource_V1NothingToCommit(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	commitDomainMock(mock, "v1")
	mock.On(v1CommitPath, func(w http.ResponseWriter, _ *http.Request, _ []byte) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"detail":"nothing to commit"}`))
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config: commitConfig(mock.URL(), "CHG-1", "a"),
			Check:  resource.TestCheckResourceAttr("mcs_firewall_commit.test", "deploy_flow", "v1"),
		}},
	})
}

func TestAccFirewallCommitResource_V1Forbidden(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	commitDomainMock(mock, "v1")
	mock.On(v1CommitPath, func(w http.ResponseWriter, _ *http.Request, _ []byte) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"detail":"forbidden"}`))
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config:      commitConfig(mock.URL(), "CHG-1", "a"),
			ExpectError: regexp.MustCompile(`firewall\.direct_management`),
		}},
	})
}

const fmgInstallPath = "/api/v3/networking/fortimanager-installs/"

// fmgInstallMock answers the install with job 42, which reports result after one poll.
func fmgInstallMock(m *mockAPIServer, result string) {
	m.On(fmgInstallPath, func(w http.ResponseWriter, _ *http.Request, _ []byte) {
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"job_id":42,"networks":3}`))
	})
	var polls atomic.Int32
	m.On("/api/v3/networking/operations/42/", func(w http.ResponseWriter, _ *http.Request, _ []byte) {
		if polls.Add(1) == 1 {
			_, _ = w.Write([]byte(`{"job_id":42,"status":"RUNNING","result":"PENDING"}`))
			return
		}
		_, _ = fmt.Fprintf(w, `{"job_id":42,"status":"PROCESSED","result":%q,"message":"install done","warnings":["ADOM locked briefly"]}`, result)
	})
}

func TestAccFirewallCommitResource_V2(t *testing.T) {
	networkOperationPollInterval = 10 * time.Millisecond
	mock := newMockAPIServer()
	defer mock.Close()
	commitDomainMock(mock, "v2")
	fmgInstallMock(mock, "SUCCESS")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config: commitConfig(mock.URL(), "CHG-1", "a"),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("mcs_firewall_commit.test", "deploy_flow", "v2"),
				resource.TestCheckResourceAttr("mcs_firewall_commit.test", "job_id", "42"),
				fwCheckCall(mock, http.MethodPost, fmgInstallPath, map[string]string{
					"domain": fmt.Sprintf("%q", commitDomainUUID), "changeid": `"CHG-1"`,
				}),
				checkCallCount(mock, http.MethodPost, v1CommitPath, 0),
				func(_ *terraform.State) error {
					if n := countCalls(mock, http.MethodGet, "/api/v3/networking/operations/42/"); n < 2 {
						return fmt.Errorf("expected the job to be polled until done, got %d polls", n)
					}
					return nil
				},
			),
		}},
	})
}

func TestAccFirewallCommitResource_V2JobFails(t *testing.T) {
	networkOperationPollInterval = 10 * time.Millisecond
	mock := newMockAPIServer()
	defer mock.Close()
	commitDomainMock(mock, "v2")
	fmgInstallMock(mock, "EXCEPTION")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config:      commitConfig(mock.URL(), "CHG-1", "a"),
			ExpectError: regexp.MustCompile(`FortiManager install for domain VDOM-001 failed`),
		}},
	})
}

func TestAccFirewallCommitResource_V2NothingPending(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	commitDomainMock(mock, "v2")
	mock.On(fmgInstallPath, func(w http.ResponseWriter, _ *http.Request, _ []byte) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"detail":"Nothing waits for an install"}`))
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config: commitConfig(mock.URL(), "CHG-1", "a"),
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("mcs_firewall_commit.test", "deploy_flow", "v2"),
				resource.TestCheckNoResourceAttr("mcs_firewall_commit.test", "job_id"),
			),
		}},
	})
}

func TestAccFirewallCommitResource_UnknownDomain(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	commitDomainMock(mock, "v1")

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config:      strings.Replace(commitConfig(mock.URL(), "CHG-1", "a"), `"VDOM-001"`, `"VDOM-999"`, 1),
			ExpectError: regexp.MustCompile(`no domain with exact name "VDOM-999"`),
		}},
	})
}
