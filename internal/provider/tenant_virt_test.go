package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// tvStore is a stateful single-object mock: POST/PUT/PATCH merge the request body into the
// stored object (starting from server-side defaults), GET returns it.
type tvStore struct {
	mu  sync.Mutex
	obj map[string]interface{}
}

func tvStatefulCRUD(mock *mockAPIServer, prefix string, defaults map[string]interface{}) *tvStore {
	s := &tvStore{}
	mock.On(prefix, func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		s.mu.Lock()
		defer s.mu.Unlock()
		merge := func() {
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			for k, v := range req {
				s.obj[k] = v
			}
		}
		switch r.Method {
		case http.MethodPost:
			s.obj = map[string]interface{}{}
			for k, v := range defaults {
				s.obj[k] = v
			}
			merge()
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(s.obj)
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(s.obj)
		case http.MethodPut, http.MethodPatch:
			merge()
			_ = json.NewEncoder(w).Encode(s.obj)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	})
	return s
}

// tvPaged serves a paginated list endpoint. Each page links to the next with an absolute URL
// that keeps the original query parameters; every received query string is recorded.
type tvPagedList struct {
	mu      sync.Mutex
	queries []string
}

func tvPaged(mock *mockAPIServer, prefix string, pages ...[]map[string]interface{}) *tvPagedList {
	p := &tvPagedList{}
	mock.On(prefix, func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		p.mu.Lock()
		p.queries = append(p.queries, r.URL.RawQuery)
		p.mu.Unlock()

		q := r.URL.Query()
		n, _ := strconv.Atoi(q.Get("page"))
		if n < 1 {
			n = 1
		}
		var results []map[string]interface{}
		if n <= len(pages) {
			results = pages[n-1]
		}
		var next interface{}
		if n < len(pages) {
			q.Set("page", strconv.Itoa(n+1))
			next = mock.URL() + r.URL.Path + "?" + q.Encode()
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count": 0, "next": next, "previous": nil, "results": results,
		})
	})
	return p
}

func (p *tvPagedList) check(t *testing.T, wantPages int, wantParam, wantValue string, forbidden ...string) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		p.mu.Lock()
		defer p.mu.Unlock()
		pagesSeen := map[string]bool{}
		for _, raw := range p.queries {
			for _, f := range forbidden {
				if strings.Contains(raw, f+"=") {
					return fmt.Errorf("query %q must not contain %s", raw, f)
				}
			}
			if wantParam != "" && !strings.Contains(raw, wantParam+"="+wantValue) {
				return fmt.Errorf("query %q is missing %s=%s", raw, wantParam, wantValue)
			}
			pagesSeen[raw] = true
		}
		if len(pagesSeen) < wantPages {
			return fmt.Errorf("expected %d distinct page requests, got %v", wantPages, p.queries)
		}
		return nil
	}
}

// tvBody returns the decoded body of the n-th call (0-based) with the given method and path prefix.
func tvBody(mock *mockAPIServer, method, prefix string, n int) (map[string]interface{}, error) {
	i := 0
	for _, c := range mock.Calls() {
		if c.Method != method || !strings.HasPrefix(c.Path, prefix) {
			continue
		}
		if i == n {
			var m map[string]interface{}
			if err := json.Unmarshal([]byte(c.Body), &m); err != nil {
				return nil, fmt.Errorf("decoding %s %s body %q: %w", method, c.Path, c.Body, err)
			}
			return m, nil
		}
		i++
	}
	return nil, fmt.Errorf("no %s call #%d to %s", method, n, prefix)
}

func tvCheckBody(mock *mockAPIServer, method, prefix string, n int, check func(map[string]interface{}) error) resource.TestCheckFunc {
	return func(_ *terraform.State) error {
		b, err := tvBody(mock, method, prefix, n)
		if err != nil {
			return err
		}
		return check(b)
	}
}

func tvJSON(v interface{}) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// ---------------------------------------------------------------------------
// mcs_customer
// ---------------------------------------------------------------------------

func TestAccCustomerResource_SdmContactsLifecycle(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	tvStatefulCRUD(mock, "/api/tenant/customers", map[string]interface{}{
		"id": "cust-001", "contractid": "", "sdm": nil,
		"tech_contacts": []int{}, "admin_contacts": []int{},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_customer" "test" {
  name = "Test Customer"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_customer.test", "id", "cust-001"),
					resource.TestCheckNoResourceAttr("mcs_customer.test", "sdm"),
					resource.TestCheckResourceAttr("mcs_customer.test", "contractid", ""),
					resource.TestCheckResourceAttr("mcs_customer.test", "tech_contacts.#", "0"),
					resource.TestCheckResourceAttr("mcs_customer.test", "admin_contacts.#", "0"),
					tvCheckBody(mock, http.MethodPost, "/api/tenant/customers/", 0, func(b map[string]interface{}) error {
						for _, k := range []string{"sdm", "contractid", "tech_contacts", "admin_contacts"} {
							if _, ok := b[k]; ok {
								return fmt.Errorf("unset %s must be omitted from POST body, got %s", k, tvJSON(b))
							}
						}
						return nil
					}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_customer" "test" {
  name           = "Test Customer"
  contractid     = "C-001"
  sdm            = 7
  tech_contacts  = [1, 2]
  admin_contacts = [3]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_customer.test", "sdm", "7"),
					resource.TestCheckResourceAttr("mcs_customer.test", "contractid", "C-001"),
					resource.TestCheckResourceAttr("mcs_customer.test", "tech_contacts.#", "2"),
					resource.TestCheckResourceAttr("mcs_customer.test", "tech_contacts.1", "2"),
					resource.TestCheckResourceAttr("mcs_customer.test", "admin_contacts.0", "3"),
					tvCheckBody(mock, http.MethodPut, "/api/tenant/customers/", 0, func(b map[string]interface{}) error {
						if b["sdm"] != float64(7) || b["contractid"] != "C-001" || tvJSON(b["tech_contacts"]) != "[1,2]" {
							return fmt.Errorf("unexpected PUT body %s", tvJSON(b))
						}
						return nil
					}),
				),
			},
			{
				// Omitting sdm keeps the server value; an explicit [] clears a contact list.
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_customer" "test" {
  name           = "Test Customer"
  tech_contacts  = []
  admin_contacts = [3]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_customer.test", "sdm", "7"),
					resource.TestCheckResourceAttr("mcs_customer.test", "contractid", "C-001"),
					resource.TestCheckResourceAttr("mcs_customer.test", "tech_contacts.#", "0"),
					tvCheckBody(mock, http.MethodPut, "/api/tenant/customers/", 1, func(b map[string]interface{}) error {
						if tvJSON(b["tech_contacts"]) != "[]" {
							return fmt.Errorf("tech_contacts must be sent as [], got %s", tvJSON(b))
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccCustomerDataSource_Paginated(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	list := tvPaged(mock, "/api/tenant/customers",
		[]map[string]interface{}{
			{"id": "cust-001", "name": "Acme Old", "contractid": "", "sdm": nil, "admin_contacts": []int{}, "tech_contacts": []int{}},
		},
		[]map[string]interface{}{
			{"id": "cust-002", "name": "Acme", "contractid": "C-2", "sdm": 12, "admin_contacts": []int{4}, "tech_contacts": []int{5, 6}},
		},
	)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_customer" "test" {
  name = "Acme"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_customer.test", "id", "cust-002"),
					resource.TestCheckResourceAttr("data.mcs_customer.test", "sdm", "12"),
					resource.TestCheckResourceAttr("data.mcs_customer.test", "tech_contacts.#", "2"),
					list.check(t, 2, "name__icontains", "Acme"),
				),
			},
		},
	})
}

func TestAccCustomerDataSource_ListAllNullSdm(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	tvPaged(mock, "/api/tenant/customers",
		[]map[string]interface{}{
			{"id": "cust-001", "name": "A", "sdm": nil},
		},
		[]map[string]interface{}{
			{"id": "cust-002", "name": "B", "sdm": 3},
		},
	)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_customer" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_customer.all", "customers.#", "2"),
					resource.TestCheckNoResourceAttr("data.mcs_customer.all", "customers.0.sdm"),
					resource.TestCheckResourceAttr("data.mcs_customer.all", "customers.1.sdm", "3"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_contact
// ---------------------------------------------------------------------------

func TestAccContactResource_OptionalFieldsAndUpdate(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	tvStatefulCRUD(mock, "/api/tenant/contacts", map[string]interface{}{
		"id": 10, "firstname": "", "lastname": "", "email": "", "phone": "", "address": "",
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_contact" "test" {
  company = "ACME"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_contact.test", "id", "10"),
					resource.TestCheckResourceAttr("mcs_contact.test", "firstname", ""),
					resource.TestCheckResourceAttr("mcs_contact.test", "email", ""),
					tvCheckBody(mock, http.MethodPost, "/api/tenant/contacts/", 0, func(b map[string]interface{}) error {
						if _, ok := b["firstname"]; ok || b["company"] != "ACME" {
							return fmt.Errorf("unexpected POST body %s", tvJSON(b))
						}
						return nil
					}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_contact" "test" {
  company   = "ACME"
  firstname = "John"
  email     = "john@example.com"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_contact.test", "firstname", "John"),
					resource.TestCheckResourceAttr("mcs_contact.test", "email", "john@example.com"),
					resource.TestCheckResourceAttr("mcs_contact.test", "phone", ""),
					tvCheckBody(mock, http.MethodPut, "/api/tenant/contacts/10/", 0, func(b map[string]interface{}) error {
						if b["firstname"] != "John" || b["email"] != "john@example.com" {
							return fmt.Errorf("unexpected PUT body %s", tvJSON(b))
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccContactDataSource_Paginated(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	list := tvPaged(mock, "/api/tenant/contacts",
		[]map[string]interface{}{
			{"id": 10, "company": "ACME Holding", "firstname": "A"},
		},
		[]map[string]interface{}{
			{"id": 11, "company": "ACME", "firstname": "Jane", "email": "jane@acme.test"},
		},
	)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_contact" "test" {
  name = "ACME"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_contact.test", "id", "11"),
					resource.TestCheckResourceAttr("data.mcs_contact.test", "name", "ACME"),
					resource.TestCheckResourceAttr("data.mcs_contact.test", "firstname", "Jane"),
					resource.TestCheckResourceAttr("data.mcs_contact.test", "lastname", ""),
					list.check(t, 2, "company__icontains", "ACME"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_dbl
// ---------------------------------------------------------------------------

func TestAccDblResource_NewFields(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	tvStatefulCRUD(mock, "/api/dbl/", map[string]interface{}{
		"id": 5, "timestamp": "2025-01-01T00:00:00Z", "occurrence": 1, "hostname": "bad-host",
		"source": "api", "persistent": false, "blackholed": false,
		"meta": "AS1234 Example", "itsm": "", "reason": "", "labels": []interface{}{},
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
					resource.TestCheckResourceAttr("mcs_dbl.test", "source", "api"),
					resource.TestCheckResourceAttr("mcs_dbl.test", "persistent", "false"),
					resource.TestCheckResourceAttr("mcs_dbl.test", "blackholed", "false"),
					resource.TestCheckResourceAttr("mcs_dbl.test", "meta", "AS1234 Example"),
					resource.TestCheckResourceAttr("mcs_dbl.test", "itsm", ""),
					tvCheckBody(mock, http.MethodPost, "/api/dbl/", 0, func(b map[string]interface{}) error {
						for _, k := range []string{"source", "persistent", "blackholed", "itsm", "reason", "meta"} {
							if _, ok := b[k]; ok {
								return fmt.Errorf("unset %s must be omitted, got %s", k, tvJSON(b))
							}
						}
						return nil
					}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_dbl" "test" {
  ipaddress  = "10.0.0.1"
  blackholed = true
  persistent = true
  itsm       = "INC-42"
  reason     = "scanner"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_dbl.test", "blackholed", "true"),
					resource.TestCheckResourceAttr("mcs_dbl.test", "persistent", "true"),
					resource.TestCheckResourceAttr("mcs_dbl.test", "itsm", "INC-42"),
					resource.TestCheckResourceAttr("mcs_dbl.test", "reason", "scanner"),
					resource.TestCheckResourceAttr("mcs_dbl.test", "source", "api"),
					tvCheckBody(mock, http.MethodPut, "/api/dbl/10.0.0.1/", 0, func(b map[string]interface{}) error {
						if b["blackholed"] != true || b["itsm"] != "INC-42" || b["reason"] != "scanner" {
							return fmt.Errorf("unexpected PUT body %s", tvJSON(b))
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccDblResource_IpAddressRequiresReplace(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	tvStatefulCRUD(mock, "/api/dbl/", map[string]interface{}{
		"id": 5, "timestamp": "2025-01-01T00:00:00Z", "occurrence": 1, "hostname": "h",
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_dbl" "test" {
  ipaddress = "10.0.0.1"
}`,
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_dbl" "test" {
  ipaddress = "10.0.0.2"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_dbl.test", "ipaddress", "10.0.0.2"),
					func(_ *terraform.State) error {
						for _, c := range mock.Calls() {
							if c.Method == http.MethodPut {
								return fmt.Errorf("changing ipaddress must replace, not PUT %s", c.Path)
							}
						}
						deleted := false
						for _, c := range mock.Calls() {
							if c.Method == http.MethodDelete && c.Path == "/api/dbl/10.0.0.1/" {
								deleted = true
							}
						}
						if !deleted {
							return fmt.Errorf("old entry was not deleted")
						}
						return nil
					},
				),
			},
		},
	})
}

func TestAccDblDataSource_PaginatedNewFields(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	list := tvPaged(mock, "/api/dbl",
		[]map[string]interface{}{
			{"id": 1, "ipaddress": "10.0.0.1", "timestamp": "t", "occurrence": 1, "hostname": "h1", "meta": "m1"},
		},
		[]map[string]interface{}{
			{"id": 2, "ipaddress": "10.0.0.2", "timestamp": "t", "occurrence": 2, "hostname": "h2",
				"blackholed": true, "meta": "m2", "itsm": "INC-1", "reason": "r"},
		},
	)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_dbl" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_dbl.all", "dbls.#", "2"),
					resource.TestCheckNoResourceAttr("data.mcs_dbl.all", "dbls.0.blackholed"),
					resource.TestCheckResourceAttr("data.mcs_dbl.all", "dbls.1.blackholed", "true"),
					resource.TestCheckResourceAttr("data.mcs_dbl.all", "dbls.1.meta", "m2"),
					resource.TestCheckResourceAttr("data.mcs_dbl.all", "dbls.1.itsm", "INC-1"),
					resource.TestCheckResourceAttr("data.mcs_dbl.all", "dbls.1.reason", "r"),
					list.check(t, 2, "", ""),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_domain_dbl
// ---------------------------------------------------------------------------

func TestAccDomainDblResource_Labels(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	tvStatefulCRUD(mock, "/api/dbl/domaindbl", map[string]interface{}{
		"id": 3, "timestamp": "2025-01-01T00:00:00Z", "occurrence": 1,
		"source": "portal", "persistent": false, "labels": []interface{}{},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_domain_dbl" "test" {
  domainname = "evil.example.com"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_domain_dbl.test", "source", "portal"),
					resource.TestCheckResourceAttr("mcs_domain_dbl.test", "persistent", "false"),
					resource.TestCheckResourceAttr("mcs_domain_dbl.test", "labels.#", "0"),
					tvCheckBody(mock, http.MethodPost, "/api/dbl/domaindbl/", 0, func(b map[string]interface{}) error {
						if _, ok := b["source"]; ok {
							return fmt.Errorf("unset source must be omitted, got %s", tvJSON(b))
						}
						if tvJSON(b["labels"]) != "[]" {
							return fmt.Errorf("labels must be sent as [], got %s", tvJSON(b))
						}
						return nil
					}),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_domain_dbl" "test" {
  domainname = "evil.example.com"
  source     = "manual"
  labels     = ["malware", "phishing"]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_domain_dbl.test", "source", "manual"),
					resource.TestCheckResourceAttr("mcs_domain_dbl.test", "labels.#", "2"),
					resource.TestCheckResourceAttr("mcs_domain_dbl.test", "labels.1", "phishing"),
					tvCheckBody(mock, http.MethodPut, "/api/dbl/domaindbl/3/", 0, func(b map[string]interface{}) error {
						if tvJSON(b["labels"]) != `[{"name":"malware"},{"name":"phishing"}]` {
							return fmt.Errorf("unexpected labels payload %s", tvJSON(b))
						}
						return nil
					}),
				),
			},
		},
	})
}

func TestAccDomainDblDataSource_PaginatedLabels(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	list := tvPaged(mock, "/api/dbl/domaindbl",
		[]map[string]interface{}{
			{"id": 1, "domainname": "a.example.com", "timestamp": "t", "occurrence": 1, "labels": []interface{}{}},
		},
		[]map[string]interface{}{
			{"id": 2, "domainname": "b.example.com", "timestamp": "t", "occurrence": 1, "source": "feed",
				"labels": []map[string]interface{}{{"id": 9, "name": "malware"}}},
		},
	)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_domain_dbl" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_domain_dbl.all", "domain_dbls.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_domain_dbl.all", "domain_dbls.0.labels.#", "0"),
					resource.TestCheckResourceAttr("data.mcs_domain_dbl.all", "domain_dbls.1.labels.0", "malware"),
					resource.TestCheckResourceAttr("data.mcs_domain_dbl.all", "domain_dbls.1.source", "feed"),
					list.check(t, 2, "", ""),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_job data source
// ---------------------------------------------------------------------------

func TestAccJobDataSource_NewFields(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.On("/api/jobs/job", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": 43, "jobname": "deploy", "timestamp": "2025-01-01T00:00:00Z",
			"endtime": nil, "message": "", "dryrun": true, "continue_on_failure": false,
			"status":                     map[string]interface{}{"status": "running"},
			"result":                     map[string]interface{}{"result": "pending"},
			"user":                       map[string]interface{}{"id": 1, "username": "bob"},
			"can_force_continue_transit": "false",
			"parent":                     41,
			"sub_jobs":                   []int{44, 45},
			"created_at_timestamp":       "2025-01-01T00:00:00Z",
			"updated_at_timestamp":       "2025-01-01T00:01:00Z",
			"created_by_user":            1,
			"updated_by_user":            nil,
		})
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_job" "test" {
  id = 43
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_job.test", "status", "running"),
					resource.TestCheckResourceAttr("data.mcs_job.test", "result", "pending"),
					resource.TestCheckNoResourceAttr("data.mcs_job.test", "endtime"),
					resource.TestCheckResourceAttr("data.mcs_job.test", "can_force_continue_transit", "false"),
					resource.TestCheckResourceAttr("data.mcs_job.test", "parent", "41"),
					resource.TestCheckResourceAttr("data.mcs_job.test", "sub_jobs.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_job.test", "sub_jobs.1", "45"),
					resource.TestCheckResourceAttr("data.mcs_job.test", "created_at_timestamp", "2025-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("data.mcs_job.test", "updated_at_timestamp", "2025-01-01T00:01:00Z"),
					resource.TestCheckResourceAttr("data.mcs_job.test", "created_by_user", "1"),
					resource.TestCheckNoResourceAttr("data.mcs_job.test", "updated_by_user"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_virtualmachine data source
// ---------------------------------------------------------------------------

func TestAccVirtualMachineDataSource_DomainFilterPaginated(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	iface := func(vm string) []map[string]interface{} {
		return []map[string]interface{}{
			{"id": "if-" + vm, "name": "eth0", "ipaddress": "10.0.0.5", "ipv6address": "", "network": nil, "macAddress": "AA", "vm_name": vm},
		}
	}
	list := tvPaged(mock, "/api/virtualization/virtualmachine",
		[]map[string]interface{}{
			{"id": "vm-001", "name": "web-01", "cpu": 2, "memory": 2048, "os": "Linux", "state": "poweredOn", "interfaces": iface("web-01")},
		},
		[]map[string]interface{}{
			{"id": "vm-002", "name": "db-01", "cpu": 4, "memory": 4096, "os": "Linux", "state": "poweredOff", "interfaces": iface("db-01")},
		},
	)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_virtualmachine" "all" {
  domain_id = 17
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.all", "domain_id", "17"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.all", "virtual_machines.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.all", "virtual_machines.0.state", "poweredOn"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.all", "virtual_machines.1.state", "poweredOff"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.all", "virtual_machines.1.interfaces.0.vm_name", "db-01"),
					resource.TestCheckNoResourceAttr("data.mcs_virtualmachine.all", "virtual_machines.1.interfaces.0.network"),
					list.check(t, 2, "interfaces__network__domain__id", "17"),
				),
			},
		},
	})
}

func TestAccVirtualMachineDataSource_ByNameState(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	list := tvPaged(mock, "/api/virtualization/virtualmachine",
		[]map[string]interface{}{
			{"id": "vm-003", "name": "web-010", "state": "poweredOn"},
		},
		[]map[string]interface{}{
			{"id": "vm-001", "name": "web-01", "state": "suspended", "interfaces": []map[string]interface{}{
				{"id": "if-1", "name": "eth0", "ipaddress": "10.0.0.5", "ipv6address": "", "network": "net-1", "macAddress": "AA", "vm_name": "web-01"},
			}},
		},
	)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_virtualmachine" "test" {
  name = "web-01"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.test", "id", "vm-001"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.test", "state", "suspended"),
					resource.TestCheckResourceAttr("data.mcs_virtualmachine.test", "interfaces.0.vm_name", "web-01"),
					list.check(t, 2, "name__icontains", "web-01", "interfaces__network__domain__id"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_virtual_datacenter data source
// ---------------------------------------------------------------------------

func TestAccVirtualDatacenterDataSource_PaginatedCluster(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	list := tvPaged(mock, "/api/virtualization/virtualdatacenter",
		[]map[string]interface{}{
			{"id": "vdc-002", "name": "prod-dc-old", "cluster": "cl-1"},
		},
		[]map[string]interface{}{
			{"id": "vdc-001", "name": "prod-dc", "cluster": "cl-2"},
		},
	)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_virtual_datacenter" "test" {
  name = "prod-dc"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_virtual_datacenter.test", "id", "vdc-001"),
					resource.TestCheckResourceAttr("data.mcs_virtual_datacenter.test", "cluster", "cl-2"),
					list.check(t, 2, "name__icontains", "prod-dc", "name__contains"),
				),
			},
		},
	})
}

// ---------------------------------------------------------------------------
// mcs_disk data source
// ---------------------------------------------------------------------------

func TestAccDiskDataSource_PaginatedClientSideName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	list := tvPaged(mock, "/api/virtualization/disk",
		[]map[string]interface{}{
			{"id": "disk-001", "name": "sda", "size": 100, "path": "/dev/sda", "type": "thin"},
		},
		[]map[string]interface{}{
			{"id": "disk-002", "name": "sdb", "size": 200, "path": "/dev/sdb", "type": "thick"},
		},
	)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_disk" "test" {
  name = "sdb"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_disk.test", "id", "disk-002"),
					resource.TestCheckResourceAttr("data.mcs_disk.test", "type", "thick"),
					list.check(t, 2, "", "", "name__icontains"),
				),
			},
		},
	})
}
