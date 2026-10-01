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

// singleObjectStore serves a single API object under basePath: POST creates it (assigning id),
// GET/PUT/PATCH/DELETE on basePath+id+"/" read, replace/merge and delete it. onWrite lets a
// test simulate server-filled fields after every write.
type singleObjectStore struct {
	mu      sync.Mutex
	obj     map[string]interface{}
	deleted bool
	writes  []string // "METHOD path" of every write request
}

func newSingleObjectStore(mock *mockAPIServer, basePath, id string, defaults map[string]interface{}, onWrite func(obj, req map[string]interface{})) *singleObjectStore {
	s := &singleObjectStore{obj: map[string]interface{}{}}
	itemPath := basePath + id + "/"
	mock.On(basePath, func(w http.ResponseWriter, r *http.Request, body []byte) {
		s.mu.Lock()
		defer s.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path != basePath && r.URL.Path != itemPath {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"detail":"Not found."}`)
			return
		}

		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch:
			s.writes = append(s.writes, r.Method+" "+r.URL.Path)
			var req map[string]interface{}
			_ = json.Unmarshal(body, &req)
			if r.Method == http.MethodPost {
				s.obj = map[string]interface{}{"id": id}
				for k, v := range defaults {
					s.obj[k] = v
				}
				s.deleted = false
			}
			for k, v := range req {
				s.obj[k] = v
			}
			if onWrite != nil {
				onWrite(s.obj, req)
			}
			if r.Method == http.MethodPost {
				w.WriteHeader(http.StatusCreated)
			}
			_ = json.NewEncoder(w).Encode(s.obj)
		case http.MethodGet:
			if s.deleted || len(s.obj) == 0 {
				w.WriteHeader(http.StatusNotFound)
				_, _ = fmt.Fprint(w, `{"detail":"Not found."}`)
				return
			}
			_ = json.NewEncoder(w).Encode(s.obj)
		case http.MethodDelete:
			s.writes = append(s.writes, r.Method+" "+r.URL.Path)
			s.deleted = true
			w.WriteHeader(http.StatusNoContent)
		}
	})
	return s
}

// checkWrite asserts that a write with the given method and path was made.
func (s *singleObjectStore) checkWrite(method, path string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		s.mu.Lock()
		defer s.mu.Unlock()
		want := method + " " + path
		for _, w := range s.writes {
			if w == want {
				return nil
			}
		}
		return fmt.Errorf("expected request %q, got %s", want, strings.Join(s.writes, ", "))
	}
}

// checkStored asserts that the stored object has the given JSON value for key.
func (s *singleObjectStore) checkStored(key string, want interface{}) resource.TestCheckFunc {
	return func(*terraform.State) error {
		s.mu.Lock()
		defer s.mu.Unlock()
		got, ok := s.obj[key]
		if !ok {
			return fmt.Errorf("key %q not stored", key)
		}
		if fmt.Sprint(got) != fmt.Sprint(want) {
			return fmt.Errorf("stored %q = %v, want %v", key, got, want)
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// mcs_application
// ---------------------------------------------------------------------------

func TestAccApplicationResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	store := newSingleObjectStore(mock, "/api/loadbalancing/application/", "app-001",
		map[string]interface{}{"pentested": false, "pentest_type": "unknown", "pentest_date": nil, "pentest_findings": nil}, nil)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_application" "test" {
  name = "webshop"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_application.test", "id", "app-001"),
					resource.TestCheckResourceAttr("mcs_application.test", "name", "webshop"),
					resource.TestCheckResourceAttr("mcs_application.test", "pentested", "false"),
					resource.TestCheckResourceAttr("mcs_application.test", "pentest_type", "unknown"),
					resource.TestCheckNoResourceAttr("mcs_application.test", "pentest_date"),
					resource.TestCheckNoResourceAttr("mcs_application.test", "pentest_findings"),
					store.checkWrite(http.MethodPost, "/api/loadbalancing/application/"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_application" "test" {
  name             = "webshop-v2"
  pentested        = true
  pentest_type     = "graybox"
  pentest_date     = "2026-03-15"
  pentest_findings = 4
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_application.test", "id", "app-001"),
					resource.TestCheckResourceAttr("mcs_application.test", "name", "webshop-v2"),
					resource.TestCheckResourceAttr("mcs_application.test", "pentested", "true"),
					resource.TestCheckResourceAttr("mcs_application.test", "pentest_type", "graybox"),
					resource.TestCheckResourceAttr("mcs_application.test", "pentest_date", "2026-03-15"),
					resource.TestCheckResourceAttr("mcs_application.test", "pentest_findings", "4"),
					store.checkWrite(http.MethodPut, "/api/loadbalancing/application/app-001/"),
				),
			},
			{
				// Removing the nullable fields sends null so the API clears them.
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_application" "test" {
  name         = "webshop-v2"
  pentested    = true
  pentest_type = "graybox"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr("mcs_application.test", "pentest_date"),
					resource.TestCheckNoResourceAttr("mcs_application.test", "pentest_findings"),
					store.checkStored("pentest_date", nil),
				),
			},
			{
				ResourceName:      "mcs_application.test",
				ImportState:       true,
				ImportStateId:     "app-001",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccApplicationResource_InvalidPentestType(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories("http://127.0.0.1:1"),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock("http://127.0.0.1:1") + `
resource "mcs_application" "test" {
  name         = "webshop"
  pentest_type = "redteam"
}`,
				ExpectError: regexpMustCompile(`value must be one of`),
			},
		},
	})
}
