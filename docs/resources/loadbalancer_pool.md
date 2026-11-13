---
page_title: "mcs_loadbalancer_pool Resource - mcs"
subcategory: "Networking"
description: |-
  Manages a load balancer pool: a subnet allocated from an `lbpools` IP pool for a customer. Only `name` and `description` can be changed in place; changing `ip_pool`, `customer`, `prefix` or `prefix_length` replaces the pool. A pool can only be destroyed while it is `free` or assigned to a Kubernetes cluster. If it is still assigned to a load balancer, destroy fails with a clear "pool is still in use" error.
---

# mcs_loadbalancer_pool (Resource)

Manages a load balancer pool: a subnet allocated from an `lbpools` IP pool for a customer. Only `name` and `description` can be changed in place; changing `ip_pool`, `customer`, `prefix` or `prefix_length` replaces the pool. A pool can only be destroyed while it is `free` or assigned to a Kubernetes cluster. If it is still assigned to a load balancer, destroy fails with a clear "pool is still in use" error.

## Example Usage

```hcl
data "mcs_customer" "acme" {
  name = "Acme"
}

data "mcs_ippool" "lbpools" {
  name = "Loadbalancer Pool"
}

resource "mcs_loadbalancer_pool" "web" {
  name        = "web-pool"
  description = "Web frontends"
  ip_pool     = data.mcs_ippool.lbpools.id
  customer    = data.mcs_customer.acme.id
}
```

## Attributes

| Attribute     | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `name`        | String | **Yes**  | Pool name. |
| `description` | String | No       | Description. |
| `ip_pool`     | String | **Yes**  | UUID of the `lbpools` IP pool to allocate from. Changing it forces replacement. |
| `customer`    | String | **Yes**  | Customer the pool belongs to. Changing it forces replacement. |
| `prefix`      | String | No       | Specific subnet to allocate, in CIDR notation (e.g. `203.0.113.16/28`). Allocated by the API when omitted. Changing it forces replacement. |
| `prefix_length` | Number | No     | Length of the subnet to allocate (e.g. `28`). Chosen by the API when omitted; read back from `prefix`. Changing it forces replacement. |

**Read-only attributes:** `id` (String), `ip_pool_name` (String), `status` (String), `assigned_object_type` (String), `assigned_object_id` (String), `assigned_object_label` (String).


## Import

```shell
terraform import mcs_loadbalancer_pool.web <loadbalancer-pool-uuid>
```
