data "mcs_lbv_server" "web_lb" {
  name = "web-lbvserver"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_cs_action" "route_to_backend" {
  name         = "route-to-web"
  lbvserver    = data.mcs_lbv_server.web_lb.id
  customer     = data.mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
