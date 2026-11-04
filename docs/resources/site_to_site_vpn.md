---
page_title: "mcs_site_to_site_vpn Resource - mcs"
subcategory: "Networking"
description: |-
  Manages a site-to-site VPN tunnel.
---

# mcs_site_to_site_vpn (Resource)

Manages a site-to-site VPN tunnel.

## Example Usage

```hcl
data "mcs_firewall" "internet" {
  name = "Internet-FW"
}

data "mcs_contact" "admin" {
  name = "Example Corp"
}

resource "mcs_site_to_site_vpn" "office" {
  name     = "office-vpn-tunnel"
  domain   = 42
  tenant   = 7
  firewall = data.mcs_firewall.internet.id
  contact  = [data.mcs_contact.admin.id]
}
```

## Attributes

| Attribute     | Type   | Required | Description |
|--------------|--------|----------|-------------|
| `domain`     | Number | **Yes**  | Domain ID. Changing it forces a new resource. |
| `tenant`     | Number | **Yes**  | Tenant ID. Changing it forces a new resource. |
| `name`       | String | No       | VPN tunnel name (max 255). Defaults to `""`. |
| `customer`   | String | No       | Customer identifier. Computed by the API when omitted. |
| `firewall`   | String | No       | UUID of the firewall terminating the tunnel. |
| `contact`    | List(Number) | No | Contact IDs. Removing it clears the list. |
| `state`      | String | No       | Tunnel state. Computed when omitted. |
| `last_status` | String | No      | Last known status. Computed when omitted. |
| `resets`     | Number | No       | Number of tunnel resets. Computed when omitted. |
| `last_check` | String | No       | Timestamp of the last health check. Computed when omitted (may be null). |
| `last_reset` | String | No       | Timestamp of the last reset. Computed when omitted (may be null). |

**Read-only attributes:** `id` (String), `uuid` (String), `created_at_timestamp` (String), `updated_at_timestamp` (String), `created_by_user` (Number), `updated_by_user` (Number).

## Import

```shell
terraform import mcs_site_to_site_vpn.office <vpn-id>
```
