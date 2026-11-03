package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Tests for the spec update that added Customer.slug, Integration.kubernetes_endpoint,
// Domain deploy_flow/state, the Network FortiManager install fields and NetworkOperation warnings.

func TestAccCustomerResource_Slug(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	store := newSingleObjectStore(mock, "/api/tenant/customers/", "cust-001",
		map[string]interface{}{"slug": "", "contractid": "", "sdm": nil, "tech_contacts": []int{}, "admin_contacts": []int{}}, nil)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				// slug unset: not sent, and the API's empty value is adopted.
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_customer" "test" {
  name = "Acme"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_customer.test", "slug", ""),
					checkLastBodyLacks(mock, http.MethodPost, "/api/tenant/customers/", "slug"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_customer" "test" {
  name = "Acme"
  slug = "acme-zorg"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_customer.test", "slug", "acme-zorg"),
					store.checkWrite(http.MethodPut, "/api/tenant/customers/cust-001/"),
					store.checkStored("slug", "acme-zorg"),
				),
			},
			{
				ResourceName:      "mcs_customer.test",
				ImportState:       true,
				ImportStateId:     "cust-001",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCustomerResource_InvalidSlug(t *testing.T) {
	for _, slug := range []string{"Acme", "-acme", "acme-", "acme_zorg", strings.Repeat("a", 64)} {
		t.Run(slug, func(t *testing.T) {
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: testProtoV6ProviderFactories("http://127.0.0.1:1"),
				Steps: []resource.TestStep{{
					Config: providerConfigBlock("http://127.0.0.1:1") + fmt.Sprintf(`
resource "mcs_customer" "test" {
  name = "Acme"
  slug = %q
}`, slug),
					ExpectError: regexpMustCompile(`Invalid Attribute Value`),
				}},
			})
		})
	}
}

func TestAccCustomerDataSource_Slug(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	mock.On("/api/tenant/customers/", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"count": 1, "next": null, "previous": null, "results": [
			{"id": "cust-001", "name": "Acme", "slug": "acme", "contractid": "", "sdm": null, "admin_contacts": [], "tech_contacts": []}]}`)
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config: providerConfigBlock(mock.URL()) + `
data "mcs_customer" "one" {
  name = "Acme"
}
data "mcs_customer" "all" {}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.mcs_customer.one", "slug", "acme"),
				resource.TestCheckResourceAttr("data.mcs_customer.all", "customers.0.slug", "acme"),
				resource.TestCheckNoResourceAttr("data.mcs_customer.all", "slug"),
			),
		}},
	})
}

func TestAccIntegrationResource_KubernetesEndpoint(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	store := newSingleObjectStore(mock, "/api/tenant/integrations/", "5",
		map[string]interface{}{
			"type": map[string]interface{}{"name": "vSphere"}, "subtype": "", "username": "",
			"secret_expire": nil, "icon": "vsphere", "has_secret": false, "has_key": false, "kubernetes_endpoint": nil,
		},
		func(obj, _ map[string]interface{}) {
			obj["id"] = 5
			// The API fills in the sizes left unset.
			if ep, ok := obj["kubernetes_endpoint"].(map[string]interface{}); ok {
				if _, ok := ep["data_disk_max_size"]; !ok {
					ep["data_disk_max_size"] = 2000
				}
				if _, ok := ep["nodestorage_step"]; !ok {
					ep["nodestorage_step"] = 50
				}
			}
		})

	const addr = "mcs_integration.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_integration" "test" {
  name    = "vcenter01"
  type_id = 3
  baseurl = "https://vcenter01.example.org"
  kubernetes_endpoint = {
    credentials_key = "vcenter01-k8s"
    content_library = "k8s-node-images"
  }
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "kubernetes_endpoint.credentials_key", "vcenter01-k8s"),
					resource.TestCheckResourceAttr(addr, "kubernetes_endpoint.content_library", "k8s-node-images"),
					resource.TestCheckResourceAttr(addr, "kubernetes_endpoint.data_disk_max_size", "2000"),
					resource.TestCheckResourceAttr(addr, "kubernetes_endpoint.nodestorage_step", "50"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_integration" "test" {
  name    = "vcenter01"
  type_id = 3
  baseurl = "https://vcenter01.example.org"
  kubernetes_endpoint = {
    credentials_key    = "vcenter01-k8s"
    content_library    = "k8s-node-images"
    data_disk_max_size = 1000
  }
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(addr, "kubernetes_endpoint.data_disk_max_size", "1000"),
					resource.TestCheckResourceAttr(addr, "kubernetes_endpoint.nodestorage_step", "50"),
					store.checkWrite(http.MethodPut, "/api/tenant/integrations/5/"),
				),
			},
			{
				// Removing the block clears the endpoint: the update sends an explicit null.
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_integration" "test" {
  name    = "vcenter01"
  type_id = 3
  baseurl = "https://vcenter01.example.org"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckNoResourceAttr(addr, "kubernetes_endpoint.credentials_key"),
					checkLastBodyHasNull(mock, http.MethodPut, "/api/tenant/integrations/5/", "kubernetes_endpoint"),
				),
			},
		},
	})
}

func TestAccIntegrationResource_KubernetesEndpointValidation(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories("http://127.0.0.1:1"),
		Steps: []resource.TestStep{{
			Config: providerConfigBlock("http://127.0.0.1:1") + `
resource "mcs_integration" "test" {
  name    = "vcenter01"
  type_id = 3
  baseurl = "https://vcenter01.example.org"
  kubernetes_endpoint = {
    credentials_key    = "vcenter01-k8s"
    content_library    = "k8s-node-images"
    data_disk_max_size = 499
  }
}`,
			ExpectError: regexpMustCompile(`must be between 500`),
		}},
	})
}

func TestAccDomainDataSource_DeployFlowAndState(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	mock.On("/api/tenant/domains/", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"count": 2, "next": null, "previous": null, "results": [
			{"id": 1, "uuid": "d-1", "name": "dom-v1", "tag": [], "comp_errors": 0, "max_networks": 10, "deploy_flow": "v1", "state": "active"},
			{"id": 2, "uuid": "d-2", "name": "dom-v2", "tag": [], "comp_errors": 0, "max_networks": 10, "deploy_flow": "v2", "state": "planned"}]}`)
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config: providerConfigBlock(mock.URL()) + `
data "mcs_domain" "one" {
  name = "dom-v2"
}
data "mcs_domain" "all" {}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.mcs_domain.one", "deploy_flow", "v2"),
				resource.TestCheckResourceAttr("data.mcs_domain.one", "state", "planned"),
				resource.TestCheckResourceAttr("data.mcs_domain.all", "domains.0.deploy_flow", "v1"),
				resource.TestCheckResourceAttr("data.mcs_domain.all", "domains.0.state", "active"),
				resource.TestCheckNoResourceAttr("data.mcs_domain.all", "state"),
			),
		}},
	})
}

func TestNetworkV3ToModel_FortiManagerInstall(t *testing.T) {
	var n networkV3APIModel
	if err := json.Unmarshal([]byte(`{"id": "n1", "fortimanager_install_pending_since": "2026-11-01T10:00:00Z",
		"fortimanager_install_error": "policy not active"}`), &n); err != nil {
		t.Fatal(err)
	}
	m := n.toModel()
	if got := m.FmgPendingSince.ValueString(); got != "2026-11-01T10:00:00Z" {
		t.Errorf("pending_since = %q", got)
	}
	if got := m.FmgInstallError.ValueString(); got != "policy not active" {
		t.Errorf("error = %q", got)
	}

	var installed networkV3APIModel
	_ = json.Unmarshal([]byte(`{"id": "n1", "fortimanager_install_pending_since": null, "fortimanager_install_error": ""}`), &installed)
	if m := installed.toModel(); !m.FmgPendingSince.IsNull() {
		t.Errorf("pending_since should be null once installed, got %v", m.FmgPendingSince)
	}
}

func TestNetworkOperation_Warnings(t *testing.T) {
	var op networkOperationAPIModel
	if err := json.Unmarshal([]byte(`{"job_id": 1, "warnings": ["FortiManager install deferred", " "]}`), &op); err != nil {
		t.Fatal(err)
	}
	var diags diag.Diagnostics
	op.addWarnings("Network deployed with a warning", &diags)
	if diags.WarningsCount() != 1 || diags.HasError() {
		t.Fatalf("want 1 warning, got %v", diags)
	}
	if got := diags[0].Detail(); got != "FortiManager install deferred" {
		t.Errorf("detail = %q", got)
	}

	var nilOp *networkOperationAPIModel
	nilOp.addWarnings("x", &diags) // must not panic
}

// lastBody returns the JSON body of the most recent request with the given method and path.
func lastBody(mock *mockAPIServer, method, path string) (map[string]json.RawMessage, error) {
	calls := mock.Calls()
	for i := len(calls) - 1; i >= 0; i-- {
		if calls[i].Method == method && calls[i].Path == path {
			var body map[string]json.RawMessage
			if err := json.Unmarshal([]byte(calls[i].Body), &body); err != nil {
				return nil, err
			}
			return body, nil
		}
	}
	return nil, fmt.Errorf("no %s %s request was made", method, path)
}

func checkLastBodyLacks(mock *mockAPIServer, method, path, key string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		body, err := lastBody(mock, method, path)
		if err != nil {
			return err
		}
		if v, ok := body[key]; ok {
			return fmt.Errorf("%s %s sent %s = %s, want it omitted", method, path, key, v)
		}
		return nil
	}
}

func checkLastBodyHasNull(mock *mockAPIServer, method, path, key string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		body, err := lastBody(mock, method, path)
		if err != nil {
			return err
		}
		if v, ok := body[key]; !ok || string(v) != "null" {
			return fmt.Errorf("%s %s sent %s = %q, want null", method, path, key, v)
		}
		return nil
	}
}

func TestAccIntegrationDataSource_KubernetesEndpoint(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	mock.On("/api/tenant/integrations/", func(w http.ResponseWriter, r *http.Request, _ []byte) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"count": 2, "next": null, "previous": null, "results": [
			{"id": 5, "name": "vcenter01", "type": {"name": "vSphere"}, "baseurl": "https://vc", "has_secret": true, "has_key": false,
			 "kubernetes_endpoint": {"credentials_key": "vc-k8s", "content_library": "k8s-images", "data_disk_max_size": 2000, "nodestorage_step": 50}},
			{"id": 6, "name": "prtg", "type": {"name": "PRTG"}, "baseurl": "https://prtg", "has_secret": true, "has_key": false, "kubernetes_endpoint": null}]}`)
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{{
			Config: providerConfigBlock(mock.URL()) + `
data "mcs_integration" "vc" {
  name = "vcenter01"
}
data "mcs_integration" "all" {}`,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.mcs_integration.vc", "kubernetes_endpoint.credentials_key", "vc-k8s"),
				resource.TestCheckResourceAttr("data.mcs_integration.vc", "kubernetes_endpoint.content_library", "k8s-images"),
				resource.TestCheckResourceAttr("data.mcs_integration.vc", "kubernetes_endpoint.data_disk_max_size", "2000"),
				resource.TestCheckResourceAttr("data.mcs_integration.vc", "kubernetes_endpoint.nodestorage_step", "50"),
				resource.TestCheckResourceAttr("data.mcs_integration.all", "integrations.0.kubernetes_endpoint.credentials_key", "vc-k8s"),
				resource.TestCheckNoResourceAttr("data.mcs_integration.all", "integrations.1.kubernetes_endpoint.credentials_key"),
			),
		}},
	})
}
