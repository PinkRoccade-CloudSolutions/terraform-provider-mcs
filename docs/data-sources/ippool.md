---
page_title: "mcs_ippool Data Source - mcs"
subcategory: "Networking"
description: |-
  Look up IP pools. Provide `name` or `id` for a single match, or omit both to list all (optionally filtered by `customer` and `type`).
---

# mcs_ippool (Data Source)

Look up IP pools. Provide `name` or `id` for a single match, or omit both to list all (optionally filtered by `customer` and `type`).

## Example Usage

```hcl
data "mcs_ippool" "nat_pool" {
  name = "NAT pool voor MKS"
}

output "pool_id" {
  value = data.mcs_ippool.nat_pool.id
}
```

## Attributes

| Attribute  | Type   | Mode     | Description |
|-----------|--------|----------|-------------|
| `name`    | String | Optional | Exact pool name. |
| `id`      | String | Optional/Computed | Pool ID. |
| `subnet`  | String | Computed | Pool subnet. |
| `customer` | String | Optional/Computed | Filter by customer; set to the matched pool's customer. |
| `type`    | String | Optional/Computed | Filter by type (`nat`, `vip`, `loadbalancer`); set to the matched pool's type. |
| `total_ips` | String | Computed | Total number of addresses in the pool. |
| `free_ips` | String | Computed | Number of free addresses in the pool. |
| `ip_pools` | List  | Computed | List of all IP pools (populated when `name` and `id` are not set). |

**Nested `ip_pools` attributes:** `id`, `name`, `subnet`, `customer`, `type`, `total_ips`, `free_ips` — all String, Computed.
