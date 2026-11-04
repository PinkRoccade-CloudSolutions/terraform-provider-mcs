data "mcs_network_products" "all" {}

output "network_products" {
  value = data.mcs_network_products.all.names
}
