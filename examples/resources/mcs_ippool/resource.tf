data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_ippool" "nat" {
  name     = "NAT pool"
  subnet   = "203.0.113.0/28"
  type     = "nat"
  customer = data.mcs_customer.example.id
}

output "nat_pool_free_ips" {
  value = mcs_ippool.nat.free_ips
}
