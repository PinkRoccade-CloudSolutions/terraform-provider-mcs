---
page_title: "mcs_rewrite_policy Resource - mcs"
subcategory: "Load Balancing"
description: |-
  Manages a rewrite policy.
---

# mcs_rewrite_policy (Resource)

Manages a rewrite policy.

## Example Usage

```hcl
data "mcs_rewrite_action" "replace_host" {
  name = "replace-host"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_rewrite_policy" "rewrite_requests" {
  name                   = "rewrite-requests"
  rule                   = "HTTP.REQ.URL.CONTAINS(\"/old\")"
  action                 = data.mcs_rewrite_action.replace_host.id
  undefaction            = "NOREWRITE"
  comment                = "Rewrite incoming requests"
  priority               = 100
  bindpoint              = "REQUEST"
  gotopriorityexpression = "NEXT"
  customer               = data.mcs_customer.example.id
  loadbalancer           = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

## Attributes

All optional attributes except `action` are also computed: when omitted, the value set by the server is kept in state.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | **Yes** | Rewrite policy name. |
| `rule` | String | No | Rewrite rule expression. |
| `action` | String | No | Rewrite action UUID to invoke. |
| `undefaction` | String | No | Action used when the rule result is undefined (`NOREWRITE`, `RESET`, `DROP`). |
| `comment` | String | No | Rewrite policy comment. |
| `priority` | Number | No | Rewrite policy priority (integer). Lower number = higher priority. |
| `bindpoint` | String | No | Rewrite bind point: `REQUEST` or `RESPONSE`. |
| `gotopriorityexpression` | String | No | Action after this policy: `NEXT`, `END` or `USE_INVOCATION_RESULT`. |
| `customer` | String | No | Customer identifier. |
| `loadbalancer` | String | No | UUID of the load balancer. |

**Read-only attributes:** `id` (String).

## Import

```shell
terraform import mcs_rewrite_policy.rewrite_requests <rewrite-policy-uuid>
```
