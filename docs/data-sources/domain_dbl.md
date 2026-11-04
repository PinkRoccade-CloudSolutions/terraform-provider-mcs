---
page_title: "mcs_domain_dbl Data Source - mcs"
subcategory: "Deny/Block Lists"
description: |-
  Look up domain deny/block list entries. Provide `id` for a single match, or omit to list all.
---

# mcs_domain_dbl (Data Source)

Look up domain deny/block list entries. Provide `id` for a single match, or omit to list all.

## Example Usage

```hcl
data "mcs_domain_dbl" "all" {}
```

## Attributes

| Attribute    | Type   | Mode     | Description |
|-------------|--------|----------|-------------|
| `id`        | String | Optional/Computed | Entry ID. |
| `domainname` | String | Computed | Blocked domain name. |
| `timestamp` | String | Computed | Creation timestamp. |
| `source`    | String | Computed | Source of the block entry. |
| `persistent` | Bool  | Computed | Whether the entry persists. |
| `occurrence` | Number | Computed | Number of occurrences. |
| `labels`    | List(String) | Computed | Names of the attached DBL labels. |
| `domain_dbls` | List | Computed | All domain DBL entries (populated when `id` is not set); items have the same attributes. |
