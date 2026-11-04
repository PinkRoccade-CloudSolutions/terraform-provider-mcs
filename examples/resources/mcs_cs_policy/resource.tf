data "mcs_cs_action" "route_to_backend" {
  name = "route-to-backend"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_cs_policy" "by_url" {
  name         = "route-by-url"
  action       = data.mcs_cs_action.route_to_backend.id
  expression   = "HTTP.REQ.URL.PATH.STARTSWITH(\"/api\")"
  customer     = data.mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
