data "mcs_domain" "production" {
  name = "VDOM-PROD"
}

output "domain_uuid" {
  value = data.mcs_domain.production.uuid
}
