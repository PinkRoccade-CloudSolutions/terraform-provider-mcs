data "mcs_network" "production" {
  name = "001.2002.USR"
}

output "network_vlan" {
  value = data.mcs_network.production.vlan_id
}
