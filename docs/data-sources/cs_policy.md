---
page_title: "mcs_cs_policy Data Source - mcs"
subcategory: "Load Balancing"
description: |-
  Look up content switching policies. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_cs_policy (Data Source)

Look up content switching policies. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_cs_policy" "routing" {
  name = "route-by-url"
}
```

## Attributes

| Attribute      | Type   | Mode     | Description |
|---------------|--------|----------|-------------|
| `name`        | String | Optional | Exact policy name. |
| `id`          | String | Optional/Computed | Policy UUID. |
| `action`      | String | Computed | CS action name. |
| `expression`  | String | Computed | Policy expression. |
| `customer`    | String | Computed | Customer identifier. |
| `application` | String | Computed | Application identifier. |
| `loadbalancer` | String | Computed | UUID of the load balancer. |
| `cs_policies` | List   | Computed | All CS policies (populated when neither `name` nor `id` is set). |

**Nested `cs_policies` attributes:** `id`, `name`, `action`, `expression`, `customer`, `application`, `loadbalancer` — all String, Computed.
