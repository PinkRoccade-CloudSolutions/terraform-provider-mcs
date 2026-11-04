data "mcs_domain" "production" {
  name = "production"
}

data "mcs_networkpool" "lan" {
  name = "Production LAN Pool"
}

resource "mcs_network" "web" {
  domain_uuid     = data.mcs_domain.production.uuid
  product         = "DMZ"
  bitmask         = 26
  description     = "Web tier"
  network_pool_id = data.mcs_networkpool.lan.id

  timeouts {
    create = "45m"
  }
}

output "web_vlan" {
  value = mcs_network.web.vlan_id
}
