data "mcs_firewall" "internet" {
  name = "Internet-FW"
}

data "mcs_contact" "admin" {
  name = "Example Corp"
}

resource "mcs_site_to_site_vpn" "office" {
  name     = "office-vpn-tunnel"
  domain   = 42
  tenant   = 7
  firewall = data.mcs_firewall.internet.id
  contact  = [data.mcs_contact.admin.id]
}
