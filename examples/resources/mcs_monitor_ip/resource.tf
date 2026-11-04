data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_monitor_ip" "web_check" {
  ipaddress    = "203.0.113.10"
  customer     = data.mcs_customer.example.id
  notify_email = "ops@example.com"
  comment      = "Monitor production web server"
}
