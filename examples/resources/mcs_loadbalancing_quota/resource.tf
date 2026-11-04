resource "mcs_loadbalancing_quota" "standard" {
  name                      = "standard"
  max_policies_in_csvserver = 40
}
