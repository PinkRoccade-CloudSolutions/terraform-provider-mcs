---
page_title: "mcs_lb_monitor Data Source - mcs"
subcategory: "Load Balancing"
description: |-
  Look up load balancer health monitors. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_lb_monitor (Data Source)

Look up load balancer health monitors. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_lb_monitor" "http" {
  name = "http-health-check"
}
```

## Attributes

| Attribute      | Type   | Mode     | Description |
|---------------|--------|----------|-------------|
| `name`        | String | Optional | Exact monitor name. |
| `id`          | String | Optional/Computed | Monitor UUID. |
| `type`        | String | Computed | Monitor type (e.g. `HTTP`, `TCP`). |
| `interval`    | Number | Computed | Check interval in seconds. |
| `resptimeout` | Number | Computed | Response timeout in seconds. |
| `downtime`    | Number | Computed | Downtime threshold in seconds. |
| `respcode`    | String | Computed | Expected response code. |
| `secure`      | String | Computed | Secure connection setting. |
| `httprequest` | String | Computed | HTTP request string. |
| `loadbalancer` | String | Computed | UUID of the load balancer. |
| `protected`   | Bool   | Computed | Whether the monitor is protected. |
| `customer`    | String | Computed | Customer identifier. |
| `lb_monitors` | List   | Computed | All monitors (populated when neither `name` nor `id` is set). |
