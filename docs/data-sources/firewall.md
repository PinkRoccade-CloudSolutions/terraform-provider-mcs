---
page_title: "mcs_firewall Data Source - mcs"
subcategory: "Firewall"
description: |-
  Look up firewalls. Provide `name` or `id` for a single match, or omit both to list all (optionally filtered by `customer`).
---

# mcs_firewall (Data Source)

Look up firewalls. Provide `name` or `id` for a single match, or omit both to list all (optionally filtered by `customer`).

## Example Usage

```hcl
data "mcs_firewall" "internet" {
  name = "Gemnet"
}

output "firewall_id" {
  value = data.mcs_firewall.internet.id
}
```

## Attributes

| Attribute    | Type   | Mode     | Description |
|-------------|--------|----------|-------------|
| `name`      | String | Optional | Exact firewall name. |
| `id`        | String | Optional/Computed | Firewall ID (use as filter or read from result). |
| `customer`  | String | Optional/Computed | Customer filter for name and list lookups; the customer of the matched firewall. |
| `description` | String | Computed | Firewall description. |
| `customer_name` | String | Computed | Customer name. |
| `type`      | String | Computed | Firewall type (`internet` or `wan`). |
| `device`    | String | Computed | Firewall device. |
| `device_name` | String | Computed | Device name. |
| `platform`  | String | Computed | Device platform. |
| `supports_threat_protection` | Bool | Computed | Whether the device supports threat protection. |
| `context`   | String | Computed | VDOM on Fortimanager or Device Group on Panorama. |
| `external_interface` | String | Computed | External (internet or WAN facing) interface. |
| `internal_interface` | String | Computed | Internal (VDOM or transit facing) interface. |
| `default_log_profile` | String | Computed | Default log profile for new rules. |
| `default_protect_profile` | String | Computed | Default protect group profile for new rules. |
| `multi_tenant` | Bool | Computed | Whether the device is multi tenant. |
| `tag_name`  | String | Computed | TAG name for object lookups on multi tenant firewalls (PaloAlto only). |
| `nat_ip_sync_enabled` | Bool | Computed | Whether the daily Panorama NAT public IP import is enabled. |
| `firewalls` | List   | Computed | List of all firewalls (populated when `name` and `id` are not set). |

**Nested `firewalls` attributes:** all of the attributes above except `firewalls` — all Computed.
