data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_firewall" "edge" {
  customer            = data.mcs_customer.example.id
  device              = "fmg-01"
  name                = "edge-fw"
  type                = "internet"
  context             = "VDOM-PROD"
  external_interface  = "port1"
  internal_interface  = "port2"
  nat_ip_sync_enabled = false
}
