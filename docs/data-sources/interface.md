---
page_title: "mcs_interface Data Source - mcs"
subcategory: "Compute"
description: |-
  Look up network interfaces. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_interface (Data Source)

Look up network interfaces. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_interface" "eth0" {
  name = "eth0"
}

output "interface_ip" {
  value = data.mcs_interface.eth0.ipaddress
}
```

## Attributes

| Attribute    | Type   | Mode     | Description |
|-------------|--------|----------|-------------|
| `name`      | String | Optional | Exact interface name. |
| `id`        | String | Optional/Computed | Interface ID. |
| `ipaddress` | String | Computed | IPv4 address. |
| `ipv6address` | String | Computed | IPv6 address. |
| `network`   | String | Computed | Associated network. |
| `mac_address` | String | Computed | MAC address. |
| `vm_name`   | String | Computed | Name of the VM this interface belongs to. |
| `interfaces` | List  | Computed | List of all interfaces (populated when `name` and `id` are not set). |

**Nested `interfaces` attributes:** `id`, `name`, `ipaddress`, `ipv6address`, `network`, `mac_address`, `vm_name` — all String, Computed.
