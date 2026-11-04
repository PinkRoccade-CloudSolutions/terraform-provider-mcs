---
page_title: "mcs_site_to_site_vpn Data Source - mcs"
subcategory: "Networking"
description: |-
  Look up site-to-site VPN tunnels. Provide `name` (exact match, sent as the API's `name` filter) or `id` for a single match, or omit both to list all.
---

# mcs_site_to_site_vpn (Data Source)

Look up site-to-site VPN tunnels. Provide `name` (exact match, sent as the API's `name` filter) or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_site_to_site_vpn" "office" {
  name = "office-vpn-tunnel"
}
```

## Attributes

| Attribute              | Type   | Mode     | Description |
|-----------------------|--------|----------|-------------|
| `name`                | String | Optional | Exact VPN tunnel name. |
| `id`                  | String | Optional/Computed | VPN ID. |
| `uuid`                | String | Computed | VPN UUID. |
| `domain`              | Number | Computed | Domain ID. |
| `tenant`              | Number | Computed | Tenant ID. |
| `customer`            | String | Computed | Customer identifier. |
| `firewall`            | String | Computed | Firewall UUID. |
| `contact`             | List(Number) | Computed | Contact IDs. |
| `state`               | String | Computed | Tunnel state. |
| `last_status`         | String | Computed | Last known status. |
| `resets`              | Number | Computed | Number of resets. |
| `last_check`          | String | Computed | Last health check timestamp (null if never checked). |
| `last_reset`          | String | Computed | Last reset timestamp (null if never reset). |
| `created_at_timestamp` | String | Computed | Creation timestamp. |
| `updated_at_timestamp` | String | Computed | Last update timestamp. |
| `created_by_user`     | Number | Computed | ID of the user who created the VPN. |
| `updated_by_user`     | Number | Computed | ID of the user who last updated the VPN. |
| `vpns`                | List   | Computed | All VPN tunnels (populated when neither `name` nor `id` is set). |
