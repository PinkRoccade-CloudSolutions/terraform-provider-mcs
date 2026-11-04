data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_dns_domain" "example" {
  name      = "example.com"
  comment   = "Managed by Terraform"
  type      = "external"
  zone_type = "forward"
  customer  = data.mcs_customer.example.id
}

resource "mcs_dns_entry" "www" {
  domain_uuid = mcs_dns_domain.example.id
  name        = "www"
  type        = "A"
  content     = "192.0.2.1"
  expire      = 300
}
