package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func regexpMustCompile(s string) *regexp.Regexp {
	return regexp.MustCompile(s)
}

func configFromRaw(t *testing.T, s schema.Schema, values map[string]interface{}) tfsdk.Config {
	t.Helper()

	attrTypes := make(map[string]tftypes.Type, len(s.Attributes))
	for k, attr := range s.Attributes {
		attrTypes[k] = attr.GetType().TerraformType(context.Background())
	}
	objType := tftypes.Object{AttributeTypes: attrTypes}

	attrValues := make(map[string]tftypes.Value, len(s.Attributes))
	for k, attrType := range attrTypes {
		v, ok := values[k]
		if !ok || v == nil {
			attrValues[k] = tftypes.NewValue(attrType, nil)
		} else {
			attrValues[k] = tftypes.NewValue(attrType, v)
		}
	}

	raw := tftypes.NewValue(objType, attrValues)
	return tfsdk.Config{
		Schema: s,
		Raw:    raw,
	}
}

type mockRoute struct {
	prefix  string
	handler func(w http.ResponseWriter, r *http.Request, body []byte)
}

type mockAPIServer struct {
	mu     sync.Mutex
	routes []mockRoute
	calls  []recordedCall
	srv    *httptest.Server
}

type recordedCall struct {
	Method string
	Path   string
	Body   string
}

func newMockAPIServer() *mockAPIServer {
	m := &mockAPIServer{}
	m.srv = httptest.NewServer(http.HandlerFunc(m.handle))
	return m
}

func (m *mockAPIServer) URL() string {
	return m.srv.URL
}

func (m *mockAPIServer) Close() {
	m.srv.Close()
}

func (m *mockAPIServer) handle(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	_ = r.Body.Close()

	m.mu.Lock()
	m.calls = append(m.calls, recordedCall{
		Method: r.Method,
		Path:   r.URL.Path,
		Body:   string(body),
	})
	m.mu.Unlock()

	path := r.URL.Path

	m.mu.Lock()
	routes := make([]mockRoute, len(m.routes))
	copy(routes, m.routes)
	m.mu.Unlock()

	for _, route := range routes {
		if strings.HasPrefix(path, route.prefix) {
			route.handler(w, r, body)
			return
		}
	}

	w.WriteHeader(http.StatusNotFound)
	_, _ = fmt.Fprintf(w, `{"detail":"Not found: %s %s"}`, r.Method, path)
}

func (m *mockAPIServer) On(prefix string, handler func(w http.ResponseWriter, r *http.Request, body []byte)) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.routes = append(m.routes, mockRoute{prefix: prefix, handler: handler})
}

func (m *mockAPIServer) OnCRUD(basePath string, resourceJSON string) {
	var nextID = 1

	m.On(basePath, func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodPost:
			var merged map[string]interface{}
			_ = json.Unmarshal([]byte(resourceJSON), &merged)
			if len(body) > 0 {
				var reqBody map[string]interface{}
				_ = json.Unmarshal(body, &reqBody)
				for k, v := range reqBody {
					merged[k] = v
				}
			}
			if _, ok := merged["id"]; !ok {
				merged["id"] = nextID
				nextID++
			}
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(merged)

		case http.MethodGet:
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, resourceJSON)

		case http.MethodPut, http.MethodPatch:
			var merged map[string]interface{}
			_ = json.Unmarshal([]byte(resourceJSON), &merged)
			if len(body) > 0 {
				var reqBody map[string]interface{}
				_ = json.Unmarshal(body, &reqBody)
				for k, v := range reqBody {
					merged[k] = v
				}
			}
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(merged)

		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)

		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func (m *mockAPIServer) Calls() []recordedCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	result := make([]recordedCall, len(m.calls))
	copy(result, m.calls)
	return result
}

func testProtoV6ProviderFactories(serverURL string) map[string]func() (tfprotov6.ProviderServer, error) {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"mcs": providerserver.NewProtocol6WithError(New("test")()),
	}
}

func providerConfigBlock(serverURL string) string {
	return fmt.Sprintf(`
provider "mcs" {
  host     = %q
  token    = "test-token"
  insecure = true
}
`, serverURL)
}

// statefulObject is a single API object served by OnStateful. Writes merge the
// request body into the stored object, so reads reflect the latest write.
type statefulObject struct {
	mu      sync.Mutex
	data    map[string]interface{}
	deleted bool
	// deleteStatus, when non-zero, makes DELETE fail with this status.
	deleteStatus int
}

func (o *statefulObject) SetDeleteStatus(status int) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.deleteStatus = status
}

func (o *statefulObject) merge(body []byte) {
	if len(body) == 0 {
		return
	}
	var req map[string]interface{}
	_ = json.Unmarshal(body, &req)
	for k, v := range req {
		o.data[k] = v
	}
}

// OnStateful serves collection POST/GET and item GET/PUT/PATCH/DELETE for one
// object under basePath (e.g. "/api/loadbalancing/application"). The id field
// of initial is preserved across writes.
func (m *mockAPIServer) OnStateful(basePath string, initial map[string]interface{}) *statefulObject {
	obj := &statefulObject{data: map[string]interface{}{}}
	for k, v := range initial {
		obj.data[k] = v
	}
	idKey := "id"
	if _, ok := initial["uuid"]; ok {
		if _, hasID := initial["id"]; !hasID {
			idKey = "uuid"
		}
	}
	id := obj.data[idKey]

	m.On(basePath, func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		obj.mu.Lock()
		defer obj.mu.Unlock()

		isItem := strings.Trim(strings.TrimPrefix(r.URL.Path, basePath), "/") != ""
		if isItem && obj.deleted {
			w.WriteHeader(http.StatusNotFound)
			_, _ = fmt.Fprint(w, `{"detail":"Not found."}`)
			return
		}
		switch {
		case r.Method == http.MethodPost && !isItem:
			obj.merge(body)
			obj.data[idKey] = id
			obj.deleted = false
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(obj.data)
		case r.Method == http.MethodGet && !isItem:
			results := []interface{}{}
			if !obj.deleted {
				results = append(results, obj.data)
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"count": len(results), "next": nil, "previous": nil, "results": results,
			})
		case r.Method == http.MethodGet:
			_ = json.NewEncoder(w).Encode(obj.data)
		case r.Method == http.MethodPut || r.Method == http.MethodPatch:
			obj.merge(body)
			obj.data[idKey] = id
			_ = json.NewEncoder(w).Encode(obj.data)
		case r.Method == http.MethodDelete && obj.deleteStatus != 0:
			w.WriteHeader(obj.deleteStatus)
			_, _ = fmt.Fprint(w, `{"detail":"delete refused"}`)
		case r.Method == http.MethodDelete:
			obj.deleted = true
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
	return obj
}

// OnList serves a read-only collection under basePath: GET basePath returns
// all items as a paginated list, GET basePath/{id}/ returns the item whose
// "id" matches.
func (m *mockAPIServer) OnList(basePath string, items []map[string]interface{}) {
	m.On(basePath, func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		id := strings.Trim(strings.TrimPrefix(r.URL.Path, basePath), "/")
		if id == "" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"count": len(items), "next": nil, "previous": nil, "results": items,
			})
			return
		}
		for _, it := range items {
			if fmt.Sprint(it["id"]) == id {
				_ = json.NewEncoder(w).Encode(it)
				return
			}
		}
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprint(w, `{"detail":"Not found."}`)
	})
}

// checkCalled returns a check that fails unless a request with the given
// method and path was received.
func (m *mockAPIServer) checkCalled(method, path string) func(*terraform.State) error {
	return func(*terraform.State) error {
		for _, c := range m.Calls() {
			if c.Method == method && c.Path == path {
				return nil
			}
		}
		return fmt.Errorf("expected %s %s to be called", method, path)
	}
}
