data "mcs_ippool" "lbpools" {
  name = "Loadbalancer Pool"
}

resource "mcs_loadbalancer_pool" "web" {
  name        = "web-pool"
  description = "Web frontends"
  ip_pool     = data.mcs_ippool.lbpools.id
}
