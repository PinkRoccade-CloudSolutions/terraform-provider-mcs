data "mcs_ippool" "nat_pool" {
  name = "NAT Pool"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_public_ip_address" "web" {
  pool        = data.mcs_ippool.nat_pool.id
  description = "Web server public IP"
  type        = "nat"
  customer    = data.mcs_customer.example.id
}
