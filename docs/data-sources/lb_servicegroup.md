---
page_title: "mcs_lb_servicegroup Data Source - mcs"
subcategory: "Load Balancing"
description: |-
  Look up load balancer service groups. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_lb_servicegroup (Data Source)

Look up load balancer service groups. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_lb_servicegroup" "backend" {
  name = "web-backend-sg"
}
```

## Attributes

| Attribute        | Type         | Mode     | Description |
|-----------------|-------------|----------|-------------|
| `name`          | String       | Optional | Exact service group name. |
| `id`            | String       | Optional/Computed | Service group UUID. |
| `type`          | String       | Computed | Service type. |
| `state`         | String       | Computed | State: `enable` or `disable`. |
| `members`       | List(String) | Computed | Member IDs. |
| `monitors`      | List(String) | Computed | Bound monitor IDs. |
| `healthmonitor` | String       | Computed | Health monitor setting. |
| `client_certificate` | String  | Computed | Certificate presented to backend servers for client authentication. |
| `cip`           | String       | Computed | Client IP header insertion: `ENABLED` or `DISABLED`. |
| `cipheader`     | String       | Computed | Client IP header name. |
| `customer`      | String       | Computed | Customer identifier. |
| `loadbalancer`  | String       | Computed | UUID of the load balancer. |
| `lb_servicegroups` | List      | Computed | All service groups (populated when neither `name` nor `id` is set). Items have the same attributes as above. |
