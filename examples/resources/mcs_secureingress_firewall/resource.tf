data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_secureingress_firewall" "edge" {
  name        = "edge-filter"
  description = "Blocks unwanted traffic"
  customer    = data.mcs_customer.example.id
  filter_json = jsonencode({
    allow = ["10.0.0.0/8"]
  })
}
