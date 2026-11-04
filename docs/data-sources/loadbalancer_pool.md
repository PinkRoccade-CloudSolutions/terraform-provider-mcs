---
page_title: "mcs_loadbalancer_pool Data Source - mcs"
subcategory: "Networking"
description: |-
  Look up load balancer pools. Provide `name` (exact match) or `id` for a single pool, or omit both to list all of them in `pools`.
---

# mcs_loadbalancer_pool (Data Source)

Look up load balancer pools. Provide `name` (exact match) or `id` for a single pool, or omit both to list all of them in `pools`.

## Example Usage

```hcl
data "mcs_loadbalancer_pool" "web" {
  name = "web-pool"
}
```

## Attributes

| Attribute               | Type   | Mode              | Description |
|-------------------------|--------|-------------------|-------------|
| `id`                    | String | Optional/Computed | Pool ID. |
| `name`                  | String | Optional/Computed | Exact pool name. |
| `description`           | String | Computed          | Description. |
| `ip_pool`               | String | Computed          | UUID of the IP pool the subnet was cut from. |
| `ip_pool_name`          | String | Computed          | Name of the IP pool. |
| `prefix`                | String | Computed          | Allocated subnet in CIDR notation. |
| `status`                | String | Computed          | `free` or `in_use`. |
| `customer`              | String | Computed          | Customer the pool belongs to. |
| `assigned_object_type`  | String | Computed          | Type of the object the pool is assigned to. |
| `assigned_object_id`    | String | Computed          | ID of the object the pool is assigned to. |
| `assigned_object_label` | String | Computed          | Label of the object the pool is assigned to. |
| `pools`                 | List   | Computed          | All pools, each with the attributes above (populated when neither `name` nor `id` is set). |
