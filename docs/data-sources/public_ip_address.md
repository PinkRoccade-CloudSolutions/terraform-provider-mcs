---
page_title: "mcs_public_ip_address Data Source - mcs"
subcategory: "Networking"
description: |-
  Look up public IP addresses. Provide `ip_address` or `id` for a single match, or omit both to list all.
---

# mcs_public_ip_address (Data Source)

Look up public IP addresses. Provide `ip_address` or `id` for a single match, or omit both to list all.

## Example Usage — Lookup by IP address

```hcl
data "mcs_public_ip_address" "web" {
  ip_address = "203.0.113.10"
}

output "web_ip_status" {
  value = data.mcs_public_ip_address.web.status
}
```

## Example Usage — Lookup by ID

```hcl
data "mcs_public_ip_address" "by_id" {
  id = "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
}
```

## Example Usage — List all

```hcl
data "mcs_public_ip_address" "all" {}

output "all_public_ips" {
  value = [for ip in data.mcs_public_ip_address.all.public_ip_addresses : ip.ip_address]
}
```

## Attributes

| Attribute             | Type   | Mode             | Description |
|----------------------|--------|------------------|-------------|
| `ip_address`         | String | Optional/Computed | Exact public IP address to look up, or the address of the matched entry. |
| `id`                 | String | Optional/Computed | UUID for direct lookup. |
| `pool`               | String | Computed          | UUID of the IP pool (set when a single address is matched). |
| `description`        | String | Computed          | Description (set when a single address is matched). |
| `status`             | String | Computed          | Status: `available`, `assigned`, or `reserved`. |
| `type`               | String | Computed          | Type: `nat`, `vip`, `loadbalancer`, or `secureingress`. |
| `customer`           | String | Computed          | Customer identifier (set when a single address is matched). |
| `public_ip_addresses` | List  | Computed          | All public IP addresses (populated when neither `ip_address` nor `id` is set). |

**Nested `public_ip_addresses` attributes:** `id`, `ip_address`, `pool`, `description`, `status`, `type`, `customer` — all String, Computed.
