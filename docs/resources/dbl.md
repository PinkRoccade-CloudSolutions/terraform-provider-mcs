---
page_title: "mcs_dbl Resource - mcs"
subcategory: "Deny/Block Lists"
description: |-
  Manages an IP address entry in the deny/block list (DBL).
---

# mcs_dbl (Resource)

Manages an IP address entry in the deny/block list (DBL).

## Example Usage

```hcl
resource "mcs_dbl" "blocked_ip" {
  ipaddress  = "192.0.2.100"
  source     = "manual"
  persistent = true
  blackholed = false
  itsm       = "INC0012345"
}
```

## Attributes

| Attribute    | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `ipaddress` | String | **Yes**  | IP address to block (max 255 characters). Changing it forces a new entry. |
| `source`    | String | No       | Source of the block entry (e.g. `manual`, max 255 characters). Computed from the API when omitted. |
| `persistent` | Bool  | No       | Keep the entry instead of expiring it after the default time. Computed from the API when omitted. |
| `blackholed` | Bool  | No       | Drop traffic from the IP early in the defense-in-depth model (logging is then not available). Computed from the API when omitted. |
| `itsm`      | String | No       | ITSM registration in which the block was requested and substantiated (max 255 characters). |
| `reason`    | String | No       | Why the IP is added when no ITSM ticket exists. |

**Read-only attributes:**

| Attribute    | Type   | Description |
|-------------|--------|-------------|
| `id`        | String | Entry ID. |
| `timestamp` | String | Creation timestamp. |
| `occurrence` | Number | Number of occurrences. |
| `hostname`  | String | Resolved hostname. |
| `meta`      | String | Server-provided metadata about the IP address. |

> **Note:** Read, update, and delete operations use the `ipaddress` as the lookup key, not the `id`.

## Import

Import by IP address, not by `id`:

```shell
terraform import mcs_dbl.blocked_ip 192.0.2.100
```
