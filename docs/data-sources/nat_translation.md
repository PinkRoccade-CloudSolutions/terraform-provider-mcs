---
page_title: "mcs_nat_translation Data Source - mcs"
subcategory: "Networking"
description: |-
  Look up NAT translations. Provide `id` for a single match, or omit to list all.
---

# mcs_nat_translation (Data Source)

Look up NAT translations. Provide `id` for a single match, or omit to list all.

## Example Usage

```hcl
data "mcs_nat_translation" "all" {}
```

## Attributes

| Attribute          | Type   | Mode     | Description |
|-------------------|--------|----------|-------------|
| `id`              | String | Optional/Computed | NAT translation UUID. |
| `public_ip`       | String | Computed | Public IP UUID (null when unset). |
| `interface`       | String | Computed | Private interface UUID (null when unset). |
| `firewall`        | String | Computed | Firewall UUID. |
| `translation`     | String | Computed | Translation description. |
| `private_ip`      | String | Computed | Private IP address. |
| `translation_type` | String | Computed | Translation type. |
| `public_port`     | Number | Computed | Public port (if port forwarding). |
| `private_port`    | Number | Computed | Private port (if port forwarding). |
| `protocol`        | String | Computed | Protocol. |
| `customer`        | String | Computed | Customer identifier. |
| `description`     | String | Computed | Description. |
| `state`           | String | Computed | Sync state: `synced`, `unsynced`, `error`, or `deleted`. |
| `enabled`         | Bool   | Computed | Whether enabled. |
| `snat_source_addresses` | List(String) | Computed | SNAT source addresses/subnets (`snat` only). |
| `snat_translated_addresses` | List(String) | Computed | SNAT translation pool addresses (`snat` only). |
| `nat_translations` | List  | Computed | All NAT translations (populated when `id` is not set). |
