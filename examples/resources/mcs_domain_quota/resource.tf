resource "mcs_domain_quota" "standard" {
  name         = "standard"
  description  = "Standard tier"
  max_rules    = 750
  max_objects  = 1500
  max_networks = 20
}
