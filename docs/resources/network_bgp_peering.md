---
page_title: "mcs_network_bgp_peering Resource - mcs"
subcategory: "Networking"
description: |-
  Switches BGP peering on for a network (one per network) through the v3 networking API.
---

# mcs_network_bgp_peering (Resource)

Switches BGP peering on for a network through the v3 networking API, so hosts on the network, such as Kubernetes nodes, can peer with the domain's FortiGate VDOM. A network has one BGP peering, so the resource is keyed by the network's UUID.

Creating the resource switches peering on; destroying it switches peering off. Every create, update and delete queues one BGP peering change job that configures the FortiGate VDOM; Terraform waits until the job has finished. If the job fails, MCS rolls the change back on the firewall, and the error shows the failed step and the rollback error.

**Accepted prefixes.** The VDOM accepts from peers on the network only the load balancer pools of the Kubernetes clusters that use it. MCS keeps this list up to date as clusters come and go; it is exposed as `accepted_prefixes` and refreshed on every plan without causing a diff.

**Clusters still using the network.** MCS refuses (HTTP 409) to switch peering off while a Kubernetes cluster that has not been destroyed uses the network. Terraform shows MCS's reason, naming the clusters. Make the clusters depend on this resource so Terraform destroys them first.

**Created with the network.** A V2 `mcs_network` can switch peering on as part of its deploy with `bgp_peering = true`. Adding this resource for such a network is safe: switching peering on is idempotent, so the create reapplies the configuration and takes it over. Use this resource to switch peering on or off for an existing network, including V1 networks.

**Concurrency.** A network that has no domain, is not active, or has a deploy or teardown queued is refused with HTTP 409; the error is reported with MCS's reason and not retried.

## Example Usage

```hcl
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
```

## Attributes

| Attribute  | Type   | Required | Description |
|------------|--------|----------|-------------|
| `network`  | String | **Yes**  | UUID of the network to switch BGP peering on for. Changing it switches peering off on the old network and on for the new one. |
| `timeouts` | Block  | No       | `create`, `update` and `delete` durations (e.g. `"30m"`). All default to 20 minutes. |

Changing only `timeouts` switches peering on again (an idempotent change that reapplies the VDOM configuration) and waits for it.

**Read-only attributes:**

| Attribute           | Type         | Description |
|---------------------|--------------|-------------|
| `id`                | String       | UUID of the network (the BGP peering has no ID of its own). |
| `neighbor_group`    | String       | FortiGate neighbor group of the network (`MCS-K8S-<vlan id>`). |
| `accepted_prefixes` | List(String) | Prefixes the VDOM accepts from peers on the network: the load balancer pools of the Kubernetes clusters using it. Empty: it accepts nothing. |

If BGP peering is switched off outside Terraform, the next refresh removes it from state and the next apply switches it on again.

## Import

```shell
terraform import mcs_network_bgp_peering.k8s <network-uuid>
```
