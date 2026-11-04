data "mcs_lb_servicegroup_member" "web1" {
  id = "11111111-1111-4111-8111-111111111111"
}

data "mcs_lb_servicegroup_member" "web2" {
  id = "22222222-2222-4222-8222-222222222222"
}

data "mcs_lb_monitor" "http_monitor" {
  name = "http-health-check"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_lb_servicegroup" "web_backend" {
  name          = "web-backend-sg"
  type          = "HTTP"
  state         = "enable"
  members       = [data.mcs_lb_servicegroup_member.web1.id, data.mcs_lb_servicegroup_member.web2.id]
  monitors      = [data.mcs_lb_monitor.http_monitor.id]
  healthmonitor = "YES"
  cip           = "ENABLED"
  cipheader     = "X-Forwarded-For"
  customer      = data.mcs_customer.example.id
  loadbalancer  = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
