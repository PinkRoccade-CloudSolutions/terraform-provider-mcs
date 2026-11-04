---
page_title: "mcs_networkpool Data Source - mcs"
subcategory: "Networking"
description: |-
  Look up network pools. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_networkpool (Data Source)

Look up network pools. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_networkpool" "lan_pool" {
  name = "Production LAN Pool"
}
```

## Attributes

| Attribute      | Type   | Mode     | Description |
|---------------|--------|----------|-------------|
| `name`        | String | Optional | Exact network pool name. |
| `id`          | String | Optional | Network pool ID for direct lookup. |
| `network`     | String | Computed | Associated network. |
| `description` | String | Computed | Description. |
| `type`        | String | Computed | Pool type: `lan`, `wan`, or `transit`. |
| `enabled`     | Bool   | Computed | Whether the pool is enabled. |
| `network_pools` | List | Computed | List of all network pools (populated when `name` and `id` are not set). |

**Nested `network_pools` attributes:** `id`, `name`, `network`, `description`, `type` (String); `enabled` (Bool) — all Computed.
