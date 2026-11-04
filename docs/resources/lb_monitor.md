---
page_title: "mcs_lb_monitor Resource - mcs"
subcategory: "Load Balancing"
description: |-
  Manages a load balancer health monitor.
---

# mcs_lb_monitor (Resource)

Manages a load balancer health monitor.

## Example Usage

```hcl
data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_lb_monitor" "http_monitor" {
  name         = "http-health-check"
  type         = "HTTP"
  interval     = 5
  resptimeout  = 2
  downtime     = 30
  respcode     = "[\"200\"]"
  httprequest  = "GET /health"
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
  customer     = data.mcs_customer.example.id
}
```

## Attributes

All optional attributes are also computed: when omitted, the value set by the server is kept in state.

| Attribute      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `name`        | String | **Yes**  | Monitor name. |
| `type`        | String | No       | Monitor type: `HTTP`, `TCP`, `ssl_bridge`, `tcp` or `udp`. |
| `interval`    | Number | No       | Check interval in seconds. |
| `resptimeout` | Number | No       | Response timeout in seconds. |
| `downtime`    | Number | No       | Downtime threshold in seconds. |
| `respcode`    | String | No       | Accepted response codes: `["200"]`, `["200", "403", "500"]` or empty. |
| `secure`      | String | No       | Use secure connections: `YES`, `NO` or empty. |
| `httprequest` | String | No       | HTTP request string (e.g. `"GET /health"`). |
| `loadbalancer` | String | No      | UUID of the load balancer. |
| `protected`   | Bool   | No       | Whether the monitor is protected. |
| `customer`    | String | No       | Customer identifier. |

**Read-only attributes:** `id` (String).

## Import

```shell
terraform import mcs_lb_monitor.http_monitor <monitor-uuid>
```
