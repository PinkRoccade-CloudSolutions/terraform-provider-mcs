data "mcs_quota" "standard" {
  name = "standard"
}

# Resolve the domain quota profile referenced by the quota profile.
data "mcs_domain_quota" "standard" {
  id = data.mcs_quota.standard.domain_quota
}
