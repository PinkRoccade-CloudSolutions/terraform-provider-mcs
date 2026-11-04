data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_lb_servicegroup_member" "web1" {
  address      = "10.0.1.10"
  port         = 8080
  servername   = "web-server-01"
  weight       = 100
  state        = "UP"
  customer     = data.mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
