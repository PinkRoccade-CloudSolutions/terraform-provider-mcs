data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_lb_monitor" "http_monitor" {
  name         = "http-health-check"
  type         = "HTTP"
  interval     = 5
  resptimeout  = 2
  downtime     = 30
  respcode     = "[\"200\"]"
  httprequest  = "GET /health"
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
  customer     = data.mcs_customer.example.id
}
