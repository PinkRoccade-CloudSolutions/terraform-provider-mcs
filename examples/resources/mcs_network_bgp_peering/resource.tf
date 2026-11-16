data "mcs_network" "k8s" {
  name = "NET-K8S"
}

# Switch BGP peering on for an existing network, so Kubernetes nodes on it can
# peer with the domain's FortiGate VDOM.
resource "mcs_network_bgp_peering" "k8s" {
  network = data.mcs_network.k8s.id
}

output "k8s_neighbor_group" {
  value = mcs_network_bgp_peering.k8s.neighbor_group
}
