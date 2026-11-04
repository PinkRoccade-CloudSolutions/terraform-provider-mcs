---
page_title: "mcs_zone Data Source - mcs"
subcategory: "Networking"
description: |-
  Look up networking zones. Provide `name` for a single match, or omit to list all zones.
---

# mcs_zone (Data Source)

Look up networking zones. Provide `name` for a single match, or omit to list all zones.

## Example Usage — Single lookup

```hcl
data "mcs_zone" "bc_zone" {
  name = "Pink Private Cloud - Business Critical - AM5"
}

output "zone_transit_vrf" {
  value = data.mcs_zone.bc_zone.transit_vrf
}

output "zone_loadbalancers" {
  value = data.mcs_zone.bc_zone.loadbalancers
}
```

## Example Usage — List all

```hcl
data "mcs_zone" "all" {}

locals {
  target_zone = [for z in data.mcs_zone.all.zones : z if z.name == "PPC MC"][0]
}
```

## Attributes

| Attribute       | Type   | Mode     | Description |
|----------------|--------|----------|-------------|
| `name`         | String | Optional | Exact zone name to look up (filtered server-side with `name__icontains`, then matched exactly). |
| `id`           | Number | Computed | Numeric zone ID. |
| `uuid`         | String | Computed | Zone UUID. |
| `description`  | String | Computed | Zone description. |
| `adom`         | String | Computed | Administrative domain. |
| `transit_vrf`  | String | Computed | Transit VRF identifier. |
| `loadbalancers` | List   | Computed | Load balancers available in the matched zone (only set when `name` is provided). |
| `edge_root_networks` | List(String) | Computed | UUIDs of the zone's edge root networks. |
| `zones`        | List   | Computed | List of all zones (populated when `name` is not set). |

**Nested `loadbalancers` attributes:**

| Attribute | Type   | Mode     | Description |
|-----------|--------|----------|-------------|
| `id`      | String | Computed | Load balancer identifier. |
| `name`    | String | Computed | Load balancer name (the device `alias`). |

**Nested `zones` attributes:** `id` (Number), `uuid`, `name`, `description`, `adom`, `transit_vrf` (String), `edge_root_networks` (List(String)) — all Computed. Each zone also contains a nested `loadbalancers` list with the same attributes as above.
