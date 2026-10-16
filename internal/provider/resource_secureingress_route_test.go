package provider

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// ---------------------------------------------------------------------------
// mcs_secureingress_route
// ---------------------------------------------------------------------------

func checkPatchBodyContains(mock *mockAPIServer, fragments ...string) resource.TestCheckFunc {
	return func(*terraform.State) error {
		for _, c := range mock.Calls() {
			if c.Method != http.MethodPatch {
				continue
			}
			all := true
			for _, f := range fragments {
				all = all && strings.Contains(c.Body, f)
			}
			if all {
				return nil
			}
		}
		return fmt.Errorf("no PATCH request contained all of %q", fragments)
	}
}

func TestAccSecureIngressRouteResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	newRouteMock(t, mock)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		CheckDestroy:             mock.checkCalled(http.MethodDelete, "/api/secureingress/route/rt-001/"),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_secureingress_route" "test" {
  cluster        = "ic-001"
  server_name    = "Shop.Example.com"
  virtualmachine = "vm-001"
  interface      = "if-001"
  port           = 8080

  settings = {
    USE_GZIP        = "no"
    ALLOWED_METHODS = "POST|GET|GET"
    USE_MODSECURITY = "yes"
  }
  secret_settings = {
    AUTH_BASIC_PASSWORD = "hunter2"
  }

  locations = [{
    path             = "/api"
    upstream_address = "10.0.0.9"
    port             = 9000
  }]

  waf_exceptions = [{
    rules  = "942100"
    path   = "/search"
    target = "ARGS:q"
    reason = "Search box"
  }]
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "id", "rt-001"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "cluster", "ic-001"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "server_name", "Shop.Example.com"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "target_kind", "vm"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "virtualmachine", "vm-001"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "interface", "if-001"),
					resource.TestCheckNoResourceAttr("mcs_secureingress_route.test", "kubernetes_cluster"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "upstream_address", "10.0.0.5"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "upstream", "http://10.0.0.5:8080"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "port", "8080"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "scheme", "http"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "path", "/"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "ssl", "true"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "settings.%", "3"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "settings.USE_GZIP", "no"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "settings.ALLOWED_METHODS", "POST|GET|GET"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "settings.USE_MODSECURITY", "yes"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "secret_settings.AUTH_BASIC_PASSWORD", "hunter2"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "secrets_set.#", "1"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "secrets_set.0", "AUTH_BASIC_PASSWORD"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.#", "1"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.0.id", "loc-1"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.0.slot", "1"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.0.path", "/api"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.0.target_kind", "address"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.0.upstream_address", "10.0.0.9"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.0.scheme", "http"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.0.port", "9000"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.0.upstream", "http://10.0.0.9:9000"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "waf_exceptions.#", "1"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "waf_exceptions.0.id", "1"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "waf_exceptions.0.kind", "structured"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "waf_exceptions.0.rules", "942100"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "waf_exceptions.0.path", "/search"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "waf_exceptions.0.path_match", "exact"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "waf_exceptions.0.target", "ARGS:q"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "waf_exceptions.0.raw", ""),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "waf_exceptions.0.reason", "Search box"),
					resource.TestCheckNoResourceAttr("mcs_secureingress_route.test", "custom_certificate"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_secureingress_route" "test" {
  cluster        = "ic-001"
  server_name    = "Shop.Example.com"
  virtualmachine = "vm-001"
  interface      = "if-001"
  port           = 8081

  settings = {
    USE_GZIP        = "no"
    ALLOWED_METHODS = "POST|GET|GET"
  }
  secret_settings = {
    AUTH_BASIC_PASSWORD = "hunter2"
  }

  locations = [
    {
      path             = "/static"
      upstream_address = "10.0.0.10"
      port             = 80
      settings         = { USE_GZIP = "yes" }
    },
    {
      path             = "/api"
      upstream_address = "10.0.0.9"
      port             = 9000
    },
  ]

  waf_exceptions = [{
    rules      = "942100"
    path       = "/search"
    path_match = "prefix"
    target     = "ARGS:q"
    reason     = "Search box and suggestions"
  }]
}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("mcs_secureingress_route.test", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					checkPatchBodyContains(mock, `"USE_MODSECURITY":null`, `"id":"loc-1"`, `"id":1`, `"port":8081`),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "port", "8081"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "settings.%", "2"),
					resource.TestCheckNoResourceAttr("mcs_secureingress_route.test", "settings.USE_MODSECURITY"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.#", "2"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.0.id", "loc-2"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.0.path", "/static"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.0.settings.USE_GZIP", "yes"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.1.id", "loc-1"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "locations.1.path", "/api"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "waf_exceptions.0.id", "1"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "waf_exceptions.0.path_match", "prefix"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "waf_exceptions.0.reason", "Search box and suggestions"),
				),
			},
			{
				ResourceName:  "mcs_secureingress_route.test",
				ImportState:   true,
				ImportStateId: "rt-001",
				ImportStateCheck: func(states []*terraform.InstanceState) error {
					if len(states) != 1 {
						return fmt.Errorf("expected 1 imported instance, got %d", len(states))
					}
					want := map[string]string{
						"id": "rt-001", "cluster": "ic-001", "virtualmachine": "vm-001", "port": "8081",
						"locations.#": "2", "waf_exceptions.#": "1", "secrets_set.0": "AUTH_BASIC_PASSWORD",
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

const testRouteCertConfig = `
resource "mcs_secureingress_route" "test" {
  cluster        = "ic-001"
  server_name    = "shop.example.com"
  virtualmachine = "vm-001"
  interface      = "if-001"
  port           = 8080
%s
}`

const testRouteCertBlock = `
  custom_certificate = {
    certificate = <<-EOT
      -----BEGIN CERTIFICATE-----
      MIIB
      -----END CERTIFICATE-----
    EOT
    private_key = <<-EOT
      -----BEGIN PRIVATE KEY-----
      MIIE
      -----END PRIVATE KEY-----
    EOT
  }
`

func TestAccSecureIngressRouteResource_Certificate(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	newRouteMock(t, mock)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + fmt.Sprintf(testRouteCertConfig, testRouteCertBlock),
				Check: resource.ComposeAggregateTestCheckFunc(
					mock.checkCalled(http.MethodPost, "/api/secureingress/route/rt-001/certificate/"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "custom_certificate.format", "PEM"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "custom_certificate.id", "cert-001"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "custom_certificate.common_name", "shop.example.com"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "custom_certificate.subject_alt_names.#", "1"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "custom_certificate.subject_alt_names.0", "shop.example.com"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "custom_certificate.issuer", "Example CA"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "custom_certificate.serial_number", "01"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "custom_certificate.fingerprint_sha256", "AB:CD"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "custom_certificate.valid_from_timestamp", "2026-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "custom_certificate.valid_to_timestamp", "2027-01-01T00:00:00Z"),
					resource.TestCheckResourceAttrSet("mcs_secureingress_route.test", "custom_certificate.certificate"),
					resource.TestCheckResourceAttrSet("mcs_secureingress_route.test", "custom_certificate.private_key"),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + fmt.Sprintf(testRouteCertConfig, ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					mock.checkCalled(http.MethodDelete, "/api/secureingress/route/rt-001/certificate/"),
					resource.TestCheckNoResourceAttr("mcs_secureingress_route.test", "custom_certificate"),
				),
			},
		},
	})
}

func TestAccSecureIngressRouteResource_NotDeployed(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	routes := newRouteMock(t, mock)
	routes.SetConflict(true)

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config:      providerConfigBlock(mock.URL()) + fmt.Sprintf(testRouteCertConfig, ""),
				ExpectError: regexpMustCompile(`Ingress cluster not deployed yet`),
			},
		},
	})
}

func TestAccSecureIngressRouteResource_SettingsNormalisation(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()
	routes := newRouteMock(t, mock)

	config := providerConfigBlock(mock.URL()) + `
resource "mcs_secureingress_route" "test" {
  cluster        = "ic-001"
  server_name    = "SHOP.example.com."
  virtualmachine = "vm-001"
  interface      = "if-001"
  port           = 8080
  settings = {
    USE_GZIP        = "no"
    ALLOWED_METHODS = "POST|GET|POST"
    USE_MODSECURITY = "yes"
  }
}`

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				// Defaults dropped and values normalised by the API must not
				// produce a diff; the test framework fails on a non-empty plan.
				Config: config,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "server_name", "SHOP.example.com."),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "settings.ALLOWED_METHODS", "POST|GET|POST"),
					resource.TestCheckResourceAttr("mcs_secureingress_route.test", "settings.USE_GZIP", "no"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
			{
				// A real change made outside Terraform is still detected.
				PreConfig: func() {
					routes.mu.Lock()
					routes.settings["USE_MODSECURITY"] = "no"
					routes.mu.Unlock()
				},
				Config:             config,
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
func TestAccSecureIngressRouteDataSource(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	mock.OnList("/api/secureingress/route", []map[string]interface{}{
		{"id": "rt-001", "cluster": "ic-001", "server_name": "shop.example.com", "target_kind": "vm",
			"virtualmachine": "vm-001", "interface": "if-001", "kubernetes_cluster": nil, "upstream_address": "10.0.0.5",
			"port": 8080, "scheme": "http", "path": "/", "ssl": true, "upstream": "http://10.0.0.5:8080",
			"settings": map[string]interface{}{}, "secrets_set": []string{}, "locations": []interface{}{}, "waf_exceptions": []interface{}{},
			"custom_certificate": map[string]interface{}{"id": "cert-001", "common_name": "shop.example.com",
				"issuer": "Example CA", "valid_to_timestamp": "2027-01-01T00:00:00Z"}},
		{"id": "rt-002", "cluster": "ic-001", "server_name": "shop.example.com.au", "target_kind": "address",
			"virtualmachine": nil, "interface": nil, "kubernetes_cluster": nil, "upstream_address": "203.0.113.7",
			"port": 443, "scheme": "https", "path": "/", "ssl": true, "upstream": "https://203.0.113.7:443",
			"custom_certificate": nil},
	})

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
data "mcs_secureingress_route" "shop" {
  cluster     = "ic-001"
  server_name = "shop.example.com"
}

data "mcs_secureingress_route" "all" {}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.#", "1"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.id", "rt-001"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.cluster", "ic-001"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.target_kind", "vm"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.virtualmachine", "vm-001"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.interface", "if-001"),
					resource.TestCheckNoResourceAttr("data.mcs_secureingress_route.shop", "routes.0.kubernetes_cluster"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.upstream_address", "10.0.0.5"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.port", "8080"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.scheme", "http"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.path", "/"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.ssl", "true"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.upstream", "http://10.0.0.5:8080"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.certificate_common_name", "shop.example.com"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.certificate_issuer", "Example CA"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.shop", "routes.0.certificate_valid_to_timestamp", "2027-01-01T00:00:00Z"),
					resource.TestCheckResourceAttr("data.mcs_secureingress_route.all", "routes.#", "2"),
					resource.TestCheckNoResourceAttr("data.mcs_secureingress_route.all", "routes.1.certificate_common_name"),
				),
			},
		},
	})
}
