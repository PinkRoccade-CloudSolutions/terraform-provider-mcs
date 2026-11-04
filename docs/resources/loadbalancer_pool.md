---
page_title: "mcs_loadbalancer_pool Resource - mcs"
subcategory: "Networking"
description: |-
  Manages a load balancer pool: a subnet allocated from an `lbpools` IP pool. Only `name` and `description` can be changed in place; setting a different `ip_pool` replaces the pool. A pool can only be destroyed while it is `free`. If it is still assigned to a load balancer, destroy fails with a clear "pool is still in use" error.
---

# mcs_loadbalancer_pool (Resource)

Manages a load balancer pool: a subnet allocated from an `lbpools` IP pool. Only `name` and `description` can be changed in place; setting a different `ip_pool` replaces the pool. A pool can only be destroyed while it is `free`. If it is still assigned to a load balancer, destroy fails with a clear "pool is still in use" error.

## Example Usage

```hcl
data "mcs_ippool" "lbpools" {
  name = "Loadbalancer Pool"
}

resource "mcs_loadbalancer_pool" "web" {
  name        = "web-pool"
  description = "Web frontends"
  ip_pool     = data.mcs_ippool.lbpools.id
}
```

## Attributes

| Attribute     | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `name`        | String | **Yes**  | Pool name. |
| `description` | String | No       | Description. |
| `ip_pool`     | String | No       | UUID of the `lbpools` IP pool to allocate from. Selected by the API when omitted. Changing it forces replacement. |

**Read-only attributes:** `id` (String), `ip_pool_name` (String), `prefix` (String), `status` (String), `customer` (String), `assigned_object_type` (String), `assigned_object_id` (String), `assigned_object_label` (String).


## Import

```shell
terraform import mcs_loadbalancer_pool.web <loadbalancer-pool-uuid>
```
