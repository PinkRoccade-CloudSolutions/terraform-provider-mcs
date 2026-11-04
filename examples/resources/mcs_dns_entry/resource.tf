data "mcs_dns_domain" "main" {
  name = "example.com"
}

resource "mcs_dns_entry" "www" {
  domain_uuid = data.mcs_dns_domain.main.domains[0].uuid
  name        = "www"
  type        = "A"
  content     = "192.0.2.1"
  expire      = 300
}

resource "mcs_dns_entry" "mail" {
  domain_uuid = data.mcs_dns_domain.main.domains[0].uuid
  name        = "mail"
  type        = "CNAME"
  content     = "mail.example.com"
  expire      = 3600
}
