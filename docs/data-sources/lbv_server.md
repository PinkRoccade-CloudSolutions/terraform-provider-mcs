---
page_title: "mcs_lbv_server Data Source - mcs"
subcategory: "Load Balancing"
description: |-
  Look up load balancer virtual servers. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_lbv_server (Data Source)

Look up load balancer virtual servers. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_lbv_server" "web_lb" {
  name = "web-lbvserver"
}
```

## Attributes

| Attribute      | Type         | Mode     | Description |
|---------------|-------------|----------|-------------|
| `name`        | String       | Optional | Exact LB vServer name. |
| `id`          | String       | Optional/Computed | LB vServer UUID. |
| `ipaddress`   | String       | Computed | UUID of the associated PublicIPAddress. |
| `port`        | Number       | Computed | Listening port. |
| `type`        | String       | Computed | Protocol type. |
| `servicegroup` | List(String) | Computed | Service group IDs. |
| `certificate` | List(String) | Computed | SSL certificate IDs. |
| `ca_certificate` | List(String) | Computed | CA certificate IDs. |
| `customer`    | String       | Computed | Customer identifier. |
| `loadbalancer` | String      | Computed | UUID of the load balancer. |
| `lbv_servers` | List         | Computed | All LB vServers (populated when neither `name` nor `id` is set). Items have the same attributes as above. |
