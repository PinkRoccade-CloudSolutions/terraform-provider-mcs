---
page_title: "mcs_secureingress_firewall Data Source - mcs"
subcategory: "Secure Ingress"
description: |-
  Look up secure ingress (XDP) firewalls. Provide `name` or `id` for a single match, or omit both to list all. `name` is sent as the `name__icontains` filter and then matched exactly.
---

# mcs_secureingress_firewall (Data Source)

Look up secure ingress (XDP) firewalls. Provide `name` or `id` for a single match, or omit both to list all. `name` is sent as the `name__icontains` filter and then matched exactly.

## Example Usage

```hcl
data "mcs_secureingress_firewall" "edge" {
  name = "edge-filter"
}
```

## Attributes

| Attribute | Type | Mode | Description |
|-----------|------|------|-------------|
| `name` | String | Optional/Computed | Exact firewall name. |
| `id` | String | Optional/Computed | Firewall UUID. |
| `description` | String | Computed | Description of the firewall. |
| `filter_json` | String (JSON) | Computed | Firewall filter definition as a JSON document. Use `jsondecode()` to read it. |
| `customer` | String | Computed | Customer identifier. |
| `created_at_timestamp` | String | Computed | Time when the firewall was created. |
| `updated_at_timestamp` | String | Computed | Time when the firewall was last updated. |
| `created_by_user` | Number | Computed | ID of the user who created the firewall. |
| `updated_by_user` | Number | Computed | ID of the user who last updated the firewall. |
| `secureingress_firewalls` | List | Computed | All secure ingress firewalls (populated when neither `name` nor `id` is set). |

**Nested `secureingress_firewalls` attributes:** `id`, `name`, `description`, `filter_json`, `customer`, `created_at_timestamp`, `updated_at_timestamp` (String), `created_by_user`, `updated_by_user` (Number) — all Computed.
