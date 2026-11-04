data "mcs_dns_domain" "all" {}

data "mcs_dns_domain" "external" {
  type = "external"
}

data "mcs_dns_domain" "search" {
  name = "example"
}

data "mcs_dns_domain" "reverse_zones" {
  zone_type   = "reverse"
  provider_id = 3
}
