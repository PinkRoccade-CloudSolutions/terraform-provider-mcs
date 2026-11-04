data "mcs_ippool" "nat_pool" {
  name = "NAT pool voor MKS"
}

output "pool_id" {
  value = data.mcs_ippool.nat_pool.id
}
