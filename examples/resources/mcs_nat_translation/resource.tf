data "mcs_public_ip_address" "web" {
  ip_address = "203.0.113.10"
}

data "mcs_virtualmachine" "webserver" {
  name = "web-server-01"
}

data "mcs_firewall" "internet" {
  name = "Internet-FW"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_nat_translation" "web_nat" {
  public_ip        = data.mcs_public_ip_address.web.id
  interface        = data.mcs_virtualmachine.webserver.interfaces[0].id
  firewall         = data.mcs_firewall.internet.id
  translation_type = "one_to_one"
  customer         = data.mcs_customer.example.id
  description      = "1:1 NAT for web server"
}
