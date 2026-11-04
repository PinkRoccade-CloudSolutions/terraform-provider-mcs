data "mcs_public_ip_address" "web" {
  ip_address = "203.0.113.10"
}

data "mcs_lb_servicegroup" "web_backend" {
  name = "web-backend-sg"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_lbv_server" "web_lb" {
  name         = "web-lbvserver"
  ipaddress    = data.mcs_public_ip_address.web.id
  port         = 443
  type         = "ssl"
  servicegroup = [data.mcs_lb_servicegroup.web_backend.id]
  customer     = data.mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
