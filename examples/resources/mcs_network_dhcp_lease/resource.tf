data "mcs_network" "office" {
  name = "NET-OFFICE"
}

resource "mcs_network_dhcp_server" "office" {
  network     = data.mcs_network.office.id
  pool_start  = "10.20.30.100"
  pool_end    = "10.20.30.200"
  dns_servers = ["10.1.1.1"]
  lease_time  = 86400
}

locals {
  reservations = {
    printer = { ip = "10.20.30.11", mac = "00:50:56:aa:bb:01" }
    scanner = { ip = "10.20.30.12", mac = "00-50-56-AA-BB-02" }
  }
}

resource "mcs_network_dhcp_lease" "office" {
  for_each = local.reservations

  # Referring to the server orders the lease after its creation and before its removal.
  network     = mcs_network_dhcp_server.office.network
  ip          = each.value.ip
  mac         = each.value.mac
  description = each.key
}
