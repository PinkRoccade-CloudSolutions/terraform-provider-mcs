---
page_title: "mcs_monitor_ip Data Source - mcs"
subcategory: "Monitoring"
description: |-
  Look up IP address monitoring entries. Provide `id` for a single match, or omit to list all.
---

# mcs_monitor_ip (Data Source)

Look up IP address monitoring entries. Provide `id` for a single match, or omit to list all.

## Example Usage

```hcl
data "mcs_monitor_ip" "all" {}
```

## Attributes

| Attribute              | Type   | Mode     | Description |
|-----------------------|--------|----------|-------------|
| `id`                  | String | Optional/Computed | Monitor entry UUID. |
| `ipaddress`           | String | Computed | Monitored IP address. |
| `timestamp`           | String | Computed | Creation timestamp. |
| `notify_email`        | String | Computed | Notification email. |
| `last_check_timestamp` | String | Computed | Last check timestamp. |
| `customer`            | String | Computed | Customer identifier. |
| `comment`             | String | Computed | Comment. |
| `monitor_ips`         | List   | Computed | All monitor entries (populated when `id` is not set). |
