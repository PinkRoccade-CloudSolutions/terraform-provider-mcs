---
page_title: "mcs_dns_domain Data Source - mcs"
subcategory: "DNS"
description: |-
  Look up DNS domains managed by MCS. Optionally filter by name, type, zone type, customer or DNS provider, or omit filters to list all.
---

# mcs_dns_domain (Data Source)

Look up DNS domains managed by MCS. Optionally filter by name, type, zone type, customer or DNS provider, or omit filters to list all.

## Example Usage

```hcl
data "mcs_dns_domain" "all" {}

data "mcs_dns_domain" "external" {
  type = "external"
}

data "mcs_dns_domain" "search" {
  name = "example"
}

data "mcs_dns_domain" "reverse_zones" {
  zone_type   = "reverse"
  provider_id = 3
}
```

## Attributes

| Attribute | Type   | Mode     | Description |
|-----------|--------|----------|-------------|
| `name`    | String | Optional | Filter by domain name (case-insensitive contains match). |
| `type`    | String | Optional | Filter by domain type: `external` or `internal`. |
| `zone_type` | String | Optional | Filter by zone type: `forward` or `reverse`. |
| `customer` | String | Optional | Filter by customer. |
| `provider_id` | Number | Optional | Filter by DNS provider integration ID. |
| `domains` | List   | Computed | List of matching DNS domains. |

Each item in `domains` has the following attributes:

| Attribute       | Type   | Description |
|----------------|--------|-------------|
| `uuid`         | String | UUID of the DNS domain. |
| `name`         | String | Full zone name (e.g. example.com). |
| `comment`      | String | Comment for the domain. |
| `enddate`      | String | End date for the domain (if known). |
| `customer`     | String | Customer associated with the domain. |
| `provider_id`  | Number | ID of the DNS provider integration. |
| `provider_name` | String | Name of the DNS provider integration. |
| `type`         | String | Domain type: `external` or `internal`. |
| `zone_type`    | String | Zone type: `forward` or `reverse`. |
