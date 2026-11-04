data "mcs_domain_quota" "standard" {
  name = "standard"
}

data "mcs_loadbalancing_quota" "standard" {
  name = "standard"
}

resource "mcs_quota" "standard" {
  name                = "standard"
  domain_quota        = data.mcs_domain_quota.standard.id
  loadbalancing_quota = data.mcs_loadbalancing_quota.standard.id
}
