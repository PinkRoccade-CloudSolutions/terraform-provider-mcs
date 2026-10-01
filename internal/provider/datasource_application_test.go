package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// twoPageListMock serves page1 at basePath and page2 at basePath?page=2 (linked via `next`).
// Every query string received is appended to queries.
func twoPageListMock(mock *mockAPIServer, basePath string, page1, page2 []map[string]interface{}, queries *[]string) {
	mock.On(basePath, func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		if queries != nil {
			*queries = append(*queries, r.URL.RawQuery)
		}
		if r.URL.Query().Get("page") == "2" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"count": len(page1) + len(page2), "next": nil, "previous": nil, "results": page2,
			})
			return
		}
		q := r.URL.Query()
		q.Set("page", "2")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"count":    len(page1) + len(page2),
			"next":     fmt.Sprintf("%s%s?%s", mock.URL(), basePath, q.Encode()),
			"previous": nil,
			"results":  page1,
		})
	})
}

// ---------------------------------------------------------------------------
// mcs_application data source
// ---------------------------------------------------------------------------

func TestAccApplicationDataSource_ByName(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	twoPageListMock(mock, "/api/loadbalancing/application/",
		[]map[string]interface{}{
			{"id": "app-000", "name": "webshop-legacy", "pentested": false, "pentest_type": "unknown", "pentest_date": nil, "pentest_findings": nil},
		},
		[]map[string]interface{}{
			{"id": "app-001", "name": "webshop", "pentested": true, "pentest_type": "whitebox", "pentest_date": "2026-01-31", "pentest_findings": 2},
		}, nil)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_application" "test" {
  name = "webshop"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_application.test", "id", "app-001"),
					resource.TestCheckResourceAttr("data.mcs_application.test", "name", "webshop"),
					resource.TestCheckResourceAttr("data.mcs_application.test", "pentested", "true"),
					resource.TestCheckResourceAttr("data.mcs_application.test", "pentest_type", "whitebox"),
					resource.TestCheckResourceAttr("data.mcs_application.test", "pentest_date", "2026-01-31"),
					resource.TestCheckResourceAttr("data.mcs_application.test", "pentest_findings", "2"),
					resource.TestCheckResourceAttr("data.mcs_application.test", "applications.#", "0"),
				),
			},
		},
	})
}

func TestAccApplicationDataSource_ListAll(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	twoPageListMock(mock, "/api/loadbalancing/application/",
		[]map[string]interface{}{
			{"id": "app-001", "name": "webshop", "pentested": true, "pentest_type": "whitebox", "pentest_date": "2026-01-31", "pentest_findings": 2},
		},
		[]map[string]interface{}{
			{"id": "app-002", "name": "portal", "pentested": false, "pentest_type": "unknown", "pentest_date": nil, "pentest_findings": nil},
		}, nil)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_application" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_application.all", "applications.#", "2"),
					resource.TestCheckResourceAttr("data.mcs_application.all", "applications.0.id", "app-001"),
					resource.TestCheckResourceAttr("data.mcs_application.all", "applications.0.pentest_findings", "2"),
					resource.TestCheckResourceAttr("data.mcs_application.all", "applications.1.name", "portal"),
					resource.TestCheckResourceAttr("data.mcs_application.all", "applications.1.pentest_type", "unknown"),
					resource.TestCheckNoResourceAttr("data.mcs_application.all", "applications.1.pentest_date"),
				),
			},
		},
	})
}
