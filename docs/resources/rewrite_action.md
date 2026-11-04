---
page_title: "mcs_rewrite_action Resource - mcs"
subcategory: "Load Balancing"
description: |-
  Manages a rewrite action.
---

# mcs_rewrite_action (Resource)

Manages a rewrite action.

## Example Usage

```hcl
data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_rewrite_action" "replace_host" {
  name              = "replace-host"
  type              = "replace"
  target            = "HTTP.REQ.HOSTNAME"
  stringbuilderexpr = "\"example.org\""
  search            = "example.com"
  comment           = "Replace host header"
  customer          = data.mcs_customer.example.id
  loadbalancer      = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

## Attributes

All optional attributes are also computed: when omitted, the value set by the server is kept in state.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | **Yes** | Rewrite action name. |
| `type` | String | No | Rewrite action type: `replace`, `replace_all`, `replace_http_res`, `insert_http_header`, `delete_http_header`, `corrupt_http_header`, `insert_before`, `insert_before_all`, `insert_after`, `insert_after_all`, `delete`, `delete_all`, `clientless_vpn_encode`, `clientless_vpn_encode_all`, `clientless_vpn_decode`, `clientless_vpn_decode_all` or `replace_sip_res`. |
| `target` | String | No | Rewrite target expression. |
| `stringbuilderexpr` | String | No | String builder expression. |
| `search` | String | No | Search expression. |
| `comment` | String | No | Rewrite action comment. |
| `customer` | String | No | Customer identifier. |
| `loadbalancer` | String | No | UUID of the load balancer. |

**Read-only attributes:** `id` (String).

## Import

```shell
terraform import mcs_rewrite_action.replace_host <rewrite-action-uuid>
```
