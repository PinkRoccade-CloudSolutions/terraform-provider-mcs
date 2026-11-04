data "mcs_zone" "bc_zone" {
  name = "Pink Private Cloud - Business Critical - AM5"
}

output "zone_transit_vrf" {
  value = data.mcs_zone.bc_zone.transit_vrf
}

output "zone_loadbalancers" {
  value = data.mcs_zone.bc_zone.loadbalancers
}
