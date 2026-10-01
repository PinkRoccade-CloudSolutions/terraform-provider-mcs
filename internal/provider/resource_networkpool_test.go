package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccNetworkPoolResource_CRUD(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	netStatefulMock(mock, "/api/networking/networkpools/", map[string]interface{}{
		"id": "np-001", "name": "", "network": "", "description": "", "type": "", "enabled": false,
	}, nil)

	res := "mcs_networkpool.test"
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_networkpool" "test" {
  name    = "LAN Pool"
  network = "10.0.0.0/8"
  type    = "lan"
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "id", "np-001"),
					resource.TestCheckResourceAttr(res, "name", "LAN Pool"),
					resource.TestCheckResourceAttr(res, "network", "10.0.0.0/8"),
					resource.TestCheckResourceAttr(res, "type", "lan"),
					resource.TestCheckResourceAttr(res, "description", ""),
					resource.TestCheckResourceAttr(res, "enabled", "false"),
					netCheckLastBody(mock, "POST", "/api/networking/networkpools", `"name":"LAN Pool"`, `"type":"lan"`),
				),
			},
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_networkpool" "test" {
  name        = "Transit Pool"
  network     = "100.64.0.0/10"
  description = "Transit networks"
  type        = "transit"
  enabled     = true
}`,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(res, "id", "np-001"),
					resource.TestCheckResourceAttr(res, "name", "Transit Pool"),
					resource.TestCheckResourceAttr(res, "description", "Transit networks"),
					resource.TestCheckResourceAttr(res, "type", "transit"),
					resource.TestCheckResourceAttr(res, "enabled", "true"),
					netCheckLastBody(mock, "PUT", "/api/networking/networkpools/np-001", `"enabled":true`, `"network":"100.64.0.0/10"`),
				),
			},
			{
				ResourceName:      res,
				ImportState:       true,
				ImportStateId:     "np-001",
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccNetworkPoolResource_Validators(t *testing.T) {
	mock := newMockAPIServer()
	defer mock.Close()

	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testProtoV6ProviderFactories(mock.URL()),
		Steps: []resource.TestStep{
			{
				Config: providerConfigBlock(mock.URL()) + `
resource "mcs_networkpool" "test" {
  type = "dmz"
}`,
				ExpectError: regexpMustCompile(`type`),
			},
		},
	})
}
