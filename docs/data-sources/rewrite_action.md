---
page_title: "mcs_rewrite_action Data Source - mcs"
subcategory: "Load Balancing"
description: |-
  Look up rewrite actions. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_rewrite_action (Data Source)

Look up rewrite actions. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_rewrite_action" "hostname_replace" {
  name = "replace-host"
}
```

## Attributes

| Attribute | Type | Mode | Description |
|-----------|------|------|-------------|
| `name` | String | Optional | Exact rewrite action name. |
| `id` | String | Optional/Computed | Rewrite action UUID. |
| `type` | String | Computed | Rewrite action type. |
| `target` | String | Computed | Rewrite target expression. |
| `stringbuilderexpr` | String | Computed | String builder expression. |
| `search` | String | Computed | Search expression. |
| `comment` | String | Computed | Rewrite action comment. |
| `customer` | String | Computed | Customer identifier. |
| `loadbalancer` | String | Computed | UUID of the load balancer. |
| `rewrite_actions` | List | Computed | All rewrite actions (populated when neither `name` nor `id` is set). |

**Nested `rewrite_actions` attributes:** `id`, `name`, `type`, `target`, `stringbuilderexpr`, `search`, `comment`, `customer`, `loadbalancer` — all String, Computed.
