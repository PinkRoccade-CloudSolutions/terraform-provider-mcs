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

// txtQuotingMock mimics MCS/PowerDNS TXT handling as observed on a live deployment: content
// that is already quoted is rejected (MCS adds the quotes), and the list returns it quoted.
type txtQuotingMock struct {
	mu      sync.Mutex
	entries []dnsEntryAPIModel
	deletes []dnsEntryAPIModel
}

func (m *txtQuotingMock) register(mock *mockAPIServer, domainUUID string) {
	base := "/api/dns/domains/" + domainUUID + "/entries/"
	mock.On(base, func(w http.ResponseWriter, r *http.Request, body []byte) {
		w.Header().Set("Content-Type", "application/json")
		m.mu.Lock()
		defer m.mu.Unlock()
		switch {
		case r.Method == http.MethodPost && r.URL.Path == base:
			var req dnsEntryAPIModel
			_ = json.Unmarshal(body, &req)
			if req.Type == "TXT" && strings.HasPrefix(req.Content, `"`) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = fmt.Fprint(w, `{"error":"PowerDNS returned HTTP 422: Parsing record content"}`)
				return
			}
			if req.Type == "TXT" {
				req.Content = `"` + req.Content + `"`
			}
			m.entries = append(m.entries, req)
			w.WriteHeader(http.StatusCreated)
			_ = json.NewEncoder(w).Encode(req)
		case r.Method == http.MethodGet && r.URL.Path == base:
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"count": len(m.entries), "next": nil, "results": m.entries})
		case r.Method == http.MethodDelete:
			var req dnsEntryAPIModel
			_ = json.Unmarshal(body, &req)
			m.deletes = append(m.deletes, req)
			name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, base), "/")
			kept := m.entries[:0]
			for _, e := range m.entries {
				if e.Name != name {
					kept = append(kept, e)
				}
			}
			m.entries = kept
			w.WriteHeader(http.StatusOK)
			_, _ = fmt.Fprint(w, `{"result":"success"}`)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func TestAccDnsEntryResource_TXTQuoting(t *testing.T) {
	const domainUUID = "dns-domain-uuid-txt"

	for _, tc := range []struct {
		name    string
		content string // as written in HCL
		want    string // expected state value
	}{
		{name: "unquoted", content: `"tfacc provider test"`, want: "tfacc provider test"},
		{name: "quoted", content: `"\"tfacc provider test\""`, want: `"tfacc provider test"`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mock := newMockAPIServer()
			defer mock.Close()
			txt := &txtQuotingMock{}
			txt.register(mock, domainUUID)

			config := providerConfigBlock(mock.URL()) + fmt.Sprintf(`
resource "mcs_dns_entry" "txt" {
  domain_uuid = %q
  name        = "tfacc-txt"
  type        = "TXT"
  content     = %s
  expire      = 300
}`, domainUUID, tc.content)

			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
				Steps: []resource.TestStep{
					{
						// The framework's post-apply plan check fails this step if Read loses
						// the record (it would then be planned for re-creation).
						Config: config,
						Check: resource.ComposeAggregateTestCheckFunc(
							resource.TestCheckResourceAttr("mcs_dns_entry.txt", "content", tc.want),
							resource.TestCheckResourceAttr("mcs_dns_entry.txt", "id", domainUUID+"/tfacc-txt/TXT/"+tc.want),
						),
					},
					{
						ResourceName:      "mcs_dns_entry.txt",
						ImportState:       true,
						ImportStateId:     domainUUID + "/tfacc-txt/TXT/" + tc.want,
						ImportStateVerify: true,
					},
				},
				CheckDestroy: func(_ *terraform.State) error {
					txt.mu.Lock()
					defer txt.mu.Unlock()
					if len(txt.deletes) != 1 {
						return fmt.Errorf("expected 1 DELETE, got %d", len(txt.deletes))
					}
					if got := txt.deletes[0].Content; got != "tfacc provider test" {
						return fmt.Errorf("DELETE sent content %q, want it unquoted", got)
					}
					if len(txt.entries) != 0 {
						return fmt.Errorf("entry still present after destroy: %+v", txt.entries)
					}
					return nil
				},
			})
		})
	}
}
