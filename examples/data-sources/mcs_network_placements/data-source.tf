data "mcs_domain" "production" {
  name = "production"
}

data "mcs_network_placements" "production" {
  domain_uuid = data.mcs_domain.production.uuid
}

output "kubernetes_datacenters" {
  value = data.mcs_network_placements.production.kubernetes[*].name
}
