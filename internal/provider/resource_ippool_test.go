package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccIPPoolResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	netStatefulMock(mock, "/api/networking/ippools/", map[string]interface{}{
		"id": "pool-001", "type": "", "customer": "acme",
	}, func(_ string, obj map[string]interface{}) {
		if obj["subnet"] == "198.51.100.0/24" {
			obj["total_ips"] = 256
			obj["free_ips"] = 250
		} else {
			obj["total_ips"] = "16"
			obj["free_ips"] = "16"
		}
	})

	res := "mcs_ippool.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_ippool" "test" {
  name   = "Public Pool"
  subnet = "203.0.113.0/28"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "id", "pool-001"),
					resource.TestCheckResourceAttr(res, "name", "Public Pool"),
					resource.TestCheckResourceAttr(res, "subnet", "203.0.113.0/28"),
					resource.TestCheckResourceAttr(res, "type", ""),
					resource.TestCheckResourceAttr(res, "customer", "acme"),
					resource.TestCheckResourceAttr(res, "total_ips", "16"),
					resource.TestCheckResourceAttr(res, "free_ips", "16"),
					netCheckLastBody(mock, "POST", "/api/networking/ippools", `"name":"Public Pool"`, `"subnet":"203.0.113.0/28"`),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_ippool" "test" {
  name   = "NAT Pool"
  subnet = "203.0.113.0/28"
  type   = "nat"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "name", "NAT Pool"),
					resource.TestCheckResourceAttr(res, "type", "nat"),
					netCheckCallCount(mock, "POST", "/api/networking/ippools", 1),
					netCheckLastBody(mock, "PUT", "/api/networking/ippools/pool-001", `"type":"nat"`, `"customer":"acme"`),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_ippool" "test" {
  name   = "NAT Pool"
  subnet = "198.51.100.0/24"
  type   = "nat"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "subnet", "198.51.100.0/24"),
					resource.TestCheckResourceAttr(res, "total_ips", "256"),
					resource.TestCheckResourceAttr(res, "free_ips", "250"),
					netCheckCallCount(mock, "POST", "/api/networking/ippools", 2),
					netCheckCallCount(mock, "DELETE", "/api/networking/ippools", 1),
				),
			},
			{
				ResourceName:      res,
				ImportState:       true,
				ImportStateId:     "pool-001",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIPPoolResource_Validators(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_ippool" "test" {
  name   = "x"
  subnet = "10.0.0.0/8"
  type   = "public"
}`,
				ExpectError: regexpMustCompile(`type`),
			},
		},
	})
}
