---
page_title: "mcs_dbl Data Source - mcs"
subcategory: "Deny/Block Lists"
description: |-
  Look up deny/block list entries. Provide `ipaddress` for a single match, or omit to list all.
---

# mcs_dbl (Data Source)

Look up deny/block list entries. Provide `ipaddress` for a single match, or omit to list all.

## Example Usage

```hcl
data "mcs_dbl" "blocked" {
  ipaddress = "192.0.2.100"
}
```

## Attributes

| Attribute    | Type   | Mode     | Description |
|-------------|--------|----------|-------------|
| `ipaddress` | String | Optional/Computed | IP address to look up. |
| `id`        | String | Computed | Entry ID. |
| `timestamp` | String | Computed | Creation timestamp. |
| `source`    | String | Computed | Source of the block entry. |
| `occurrence` | Number | Computed | Number of occurrences. |
| `persistent` | Bool  | Computed | Whether the entry persists. |
| `blackholed` | Bool  | Computed | Whether traffic from the IP is dropped early (blackholed). |
| `hostname`  | String | Computed | Resolved hostname. |
| `meta`      | String | Computed | Server-provided metadata about the IP address. |
| `itsm`      | String | Computed | ITSM registration for the block. |
| `reason`    | String | Computed | Reason for the block when no ITSM ticket exists. |
| `dbls`      | List   | Computed | All DBL entries (populated when `ipaddress` is not set); items have the same attributes. |
