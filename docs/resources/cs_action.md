---
page_title: "mcs_cs_action Resource - mcs"
subcategory: "Load Balancing"
description: |-
  Manages a content switching action.
---

# mcs_cs_action (Resource)

Manages a content switching action.

## Example Usage

```hcl
data "mcs_lbv_server" "web_lb" {
  name = "web-lbvserver"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_cs_action" "route_to_backend" {
  name         = "route-to-web"
  lbvserver    = data.mcs_lbv_server.web_lb.id
  customer     = data.mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

## Attributes

| Attribute      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `name`        | String | **Yes**  | Action name. |
| `lbvserver`   | String | No       | UUID of the target LB vServer. |
| `customer`    | String | No       | Customer identifier. Computed by the server when omitted. |
| `loadbalancer` | String | No      | UUID of the load balancer. Computed by the server when omitted. |

**Read-only attributes:** `id` (String).

## Import

```shell
terraform import mcs_cs_action.route_to_backend <csaction-uuid>
```
