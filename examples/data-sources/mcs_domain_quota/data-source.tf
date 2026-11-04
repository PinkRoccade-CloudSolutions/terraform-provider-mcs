data "mcs_domain_quota" "standard" {
  name = "standard"
}

output "max_firewall_rules" {
  value = data.mcs_domain_quota.standard.max_rules
}
