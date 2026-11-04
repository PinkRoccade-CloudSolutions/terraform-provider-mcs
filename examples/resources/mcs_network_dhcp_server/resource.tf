data "mcs_network" "office" {
  name = "NET-OFFICE"
}

resource "mcs_network_dhcp_server" "office" {
  network     = data.mcs_network.office.id
  pool_start  = "10.20.30.100"
  pool_end    = "10.20.30.200"
  dns_servers = ["10.1.1.1", "10.1.1.2"]
  lease_time  = 86400
}

# Refer to the server's `network`, not the network itself, so Terraform creates
# leases after the server and destroys them before it (MCS refuses to disable a
# DHCP server that still has static leases).
resource "mcs_network_dhcp_lease" "printer" {
  network     = mcs_network_dhcp_server.office.network
  ip          = "10.20.30.11"
  mac         = "00:50:56:aa:bb:01"
  description = "Printer 2nd floor"
}
