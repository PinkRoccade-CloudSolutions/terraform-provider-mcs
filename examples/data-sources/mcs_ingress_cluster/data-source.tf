data "mcs_ingress_cluster" "webshop" {
  name = "webshop"
}

data "mcs_ingress_cluster" "gold" {
  sla = "gold"
}
