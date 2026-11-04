---
page_title: "mcs_ingress_cluster Resource - mcs"
subcategory: "Secure Ingress"
description: |-
  Manages a secure ingress cluster: a public IP address protected by a secure ingress (XDP) firewall and fronted by a reverse proxy. Instances, routes and upstream targets are not managed by this resource.
---

# mcs_ingress_cluster (Resource)

Manages a secure ingress cluster: a public IP address protected by a secure ingress (XDP) firewall and fronted by a reverse proxy. Instances, routes and upstream targets are not managed by this resource.

## Example Usage

```hcl
data "mcs_customer" "example" {
  name = "Example Corp"
}

data "mcs_secureingress_firewall" "edge" {
  name = "edge-xdp"
}

resource "mcs_public_ip_address" "ingress" {
  type     = "secureingress"
  customer = data.mcs_customer.example.id
}

resource "mcs_ingress_cluster" "webshop" {
  name      = "webshop"
  sla       = "gold"
  bandwidth = 1000
  customer  = data.mcs_customer.example.id
  ipaddress = mcs_public_ip_address.ingress.id
  firewall  = data.mcs_secureingress_firewall.edge.id
}
```

## Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | **Yes** | Name of the ingress cluster (max 200 characters). |
| `customer` | String | **Yes** | Customer identifier. Changing this forces a new resource. |
| `ipaddress` | String | **Yes** | UUID of the public IP address (`mcs_public_ip_address`) the cluster listens on. |
| `firewall` | String | **Yes** | UUID of the secure ingress firewall (`mcs_secureingress_firewall`). |
| `sla` | String | No | Service level: `bronze`, `silver`, `gold` or `platinum`. Server default when unset. |
| `bandwidth` | Number | No | Bandwidth in Mbps: `50`, `100`, `500`, `1000` or `5000`. Server default when unset. |

**Read-only attributes:** `id`, `slug`, `state`, `reverse_proxy_name`, `ipaddress_address`, `ipaddress_type`, `created_at_timestamp`, `updated_at_timestamp` (String), `reverse_proxy`, `created_by_user`, `updated_by_user` (Number).

## Import

```shell
terraform import mcs_ingress_cluster.webshop <ingress-cluster-uuid>
```
