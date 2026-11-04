data "mcs_customer" "example" {
  name = "Example Corp"
}

data "mcs_secureingress_firewall" "edge" {
  name = "edge-xdp"
}

resource "mcs_public_ip_address" "ingress" {
  type     = "secureingress"
  customer = data.mcs_customer.example.id
}

resource "mcs_ingress_cluster" "webshop" {
  name      = "webshop"
  sla       = "gold"
  bandwidth = 1000
  customer  = data.mcs_customer.example.id
  ipaddress = mcs_public_ip_address.ingress.id
  firewall  = data.mcs_secureingress_firewall.edge.id
}
