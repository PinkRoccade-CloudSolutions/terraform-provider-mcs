data "mcs_network" "office" {
  name = "NET-OFFICE"
}

data "mcs_network_dhcp_leases" "office" {
  network = data.mcs_network.office.id
}

output "reserved_ips" {
  value = { for l in data.mcs_network_dhcp_leases.office.leases : l.mac => l.ip }
}
