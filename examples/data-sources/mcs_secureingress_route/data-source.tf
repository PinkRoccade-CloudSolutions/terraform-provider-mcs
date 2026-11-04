data "mcs_ingress_cluster" "edge" {
  name = "edge"
}

data "mcs_secureingress_route" "shop" {
  cluster     = data.mcs_ingress_cluster.edge.id
  server_name = "shop.example.com"
}
