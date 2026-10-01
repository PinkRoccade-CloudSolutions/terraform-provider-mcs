package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccImport_MalformedCompositeIDs(t *testing.T) {
	cases := []struct {
		name, config, id, want string
	}{
		{"firewall_object", `
resource "mcs_firewall_object" "test" {
  domain = "d"
  name   = "n"
}`, "no-slash", `<domain>/<name>`},
		{"firewall_rule", `
resource "mcs_firewall_rule" "test" {
  domain = "d"
}`, "d/not-a-number", `<domain>/<policyid>`},
		{"dns_entry", `
resource "mcs_dns_entry" "test" {
  domain_uuid = "u"
  name        = "www"
  type        = "A"
  content     = "192.0.2.1"
}`, "u/www/A", `<domain_uuid>/<name>/<type>/<content>`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testProtoV6ProviderFactories("http://127.0.0.1:1"),
				Steps: []resource.TestStep{
					{
						Config:        providerConfigBlock("http://127.0.0.1:1") + c.config,
						ResourceName:  "mcs_" + c.name + ".test",
						ImportState:   true,
						ImportStateId: c.id,
						ExpectError:   regexp.MustCompile(regexp.QuoteMeta(c.want)),
					},
				},
			})
		})
	}
}

// A TXT record's content may contain slashes; the import ID keeps them in its last part.
func TestAccDnsEntryResource_ImportContentWithSlash(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	const domainUUID = "dns-domain-uuid-3"
	const content = "v=DMARC1; p=none; rua=mailto:dmarc@example.com/reports"

	mock.On("/api/dns/domains/"+domainUUID+"/entries", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]dnsEntryAPIModel{
			{Name: "_dmarc", Type: "TXT", Content: "other", Expire: 60},
			{Name: "_dmarc", Type: "TXT", Content: content, Expire: 3600},
		})
	})

	importID := domainUUID + "/_dmarc/TXT/" + content
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + fmt.Sprintf(`
resource "mcs_dns_entry" "test" {
  domain_uuid = %q
  name        = "_dmarc"
  type        = "TXT"
  content     = %q
}`, domainUUID, content),
				ResourceName:  "mcs_dns_entry.test",
				ImportState:   true,
				ImportStateId: importID,
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported instance, got %d", len(states))
					}
					want := map[string]string{
						"id": importID, "domain_uuid": domainUUID, "name": "_dmarc",
						"type": "TXT", "content": content, "expire": "3600",
					}
					for k, v := range want {
						if got := states[0].Attributes[k]; got != v {
							return fmt.Errorf("%s: got %q, want %q", k, got, v)
						}
					}
					return nil
				},
			},
		},
	})
}
