---
page_title: "mcs_cs_action Data Source - mcs"
subcategory: "Load Balancing"
description: |-
  Look up content switching actions. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_cs_action (Data Source)

Look up content switching actions. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_cs_action" "route" {
  name = "route-to-web"
}
```

## Attributes

| Attribute      | Type   | Mode     | Description |
|---------------|--------|----------|-------------|
| `name`        | String | Optional | Exact action name. |
| `id`          | String | Optional/Computed | Action UUID. |
| `lbvserver`   | String | Computed | Target LB vServer. |
| `customer`    | String | Computed | Customer identifier. |
| `loadbalancer` | String | Computed | UUID of the load balancer. |
| `cs_actions`  | List   | Computed | All CS actions (populated when neither `name` nor `id` is set). |

**Nested `cs_actions` attributes:** `id`, `name`, `lbvserver`, `customer`, `loadbalancer` — all String, Computed.
