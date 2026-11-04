data "mcs_loadbalancing_quota" "standard" {
  name = "standard"
}

output "max_policies_per_csvserver" {
  value = data.mcs_loadbalancing_quota.standard.max_policies_in_csvserver
}
