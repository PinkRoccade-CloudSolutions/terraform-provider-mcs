---
page_title: "mcs_lb_servicegroup_member Data Source - mcs"
subcategory: "Load Balancing"
description: |-
  Look up load balancer service group members. Provide `id` for a single match, or omit to list all.
---

# mcs_lb_servicegroup_member (Data Source)

Look up load balancer service group members. Provide `id` for a single match, or omit to list all.

## Example Usage

```hcl
data "mcs_lb_servicegroup_member" "all" {}
```

## Attributes

| Attribute               | Type   | Mode     | Description |
|------------------------|--------|----------|-------------|
| `id`                   | String | Optional/Computed | Member UUID. |
| `address`              | String | Computed | IP address. |
| `port`                 | Number | Computed | Port number. |
| `servername`           | String | Computed | Server name. |
| `weight`               | Number | Computed | Load balancing weight. |
| `state`                | String | Computed | Member state: `UP` or `DOWN`. |
| `customer`             | String | Computed | Customer identifier. |
| `loadbalancer`         | String | Computed | UUID of the load balancer. |
| `lb_servicegroup_members` | List | Computed | All members (populated when `id` is not set). |
