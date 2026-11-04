---
page_title: "mcs_monitor_ip Resource - mcs"
subcategory: "Monitoring"
description: |-
  Manages an IP address monitoring entry.
---

# mcs_monitor_ip (Resource)

Manages an IP address monitoring entry.

## Example Usage

```hcl
data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_monitor_ip" "web_check" {
  ipaddress    = "203.0.113.10"
  customer     = data.mcs_customer.example.id
  notify_email = "ops@example.com"
  comment      = "Monitor production web server"
}
```

## Attributes

| Attribute      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `ipaddress`   | String | **Yes**  | IP address to monitor (max 255). |
| `customer`    | String | **Yes**  | Customer identifier. |
| `notify_email` | String | No      | Comma-separated email addresses for notifications (max 255). Defaults to `""`. |
| `comment`     | String | No       | Comment. Defaults to `""`. |

**Read-only attributes:**

| Attribute              | Type   | Description |
|-----------------------|--------|-------------|
| `id`                  | String | Monitor entry ID. |
| `timestamp`           | String | Creation timestamp. |
| `last_check_timestamp` | String | Timestamp of the last check. |

## Import

```shell
terraform import mcs_monitor_ip.web_check <monitor-ip-uuid>
```
