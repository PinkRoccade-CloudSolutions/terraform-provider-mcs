---
page_title: "mcs_rewrite_policy Data Source - mcs"
subcategory: "Load Balancing"
description: |-
  Look up rewrite policies. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_rewrite_policy (Data Source)

Look up rewrite policies. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_rewrite_policy" "request_rules" {
  name = "rewrite-requests"
}
```

## Attributes

| Attribute | Type | Mode | Description |
|-----------|------|------|-------------|
| `name` | String | Optional | Exact rewrite policy name. |
| `id` | String | Optional/Computed | Rewrite policy UUID. |
| `rule` | String | Computed | Rewrite rule expression. |
| `action` | String | Computed | Rewrite action UUID. |
| `undefaction` | String | Computed | Action used when the rule result is undefined. |
| `comment` | String | Computed | Rewrite policy comment. |
| `priority` | Number (integer) | Computed | Rewrite policy priority. |
| `bindpoint` | String | Computed | Rewrite bind point. |
| `gotopriorityexpression` | String | Computed | Priority expression after evaluation. |
| `customer` | String | Computed | Customer identifier. |
| `loadbalancer` | String | Computed | UUID of the load balancer. |
| `rewrite_policies` | List | Computed | All rewrite policies (populated when neither `name` nor `id` is set). |

**Nested `rewrite_policies` attributes:** `id`, `name`, `rule`, `action`, `undefaction`, `comment`, `bindpoint`, `gotopriorityexpression`, `customer`, `loadbalancer` (String) and `priority` (Number) — all Computed.
