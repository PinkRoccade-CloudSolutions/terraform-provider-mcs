---
page_title: "mcs_cs_policy Resource - mcs"
subcategory: "Load Balancing"
description: |-
  Manages a content switching policy.
---

# mcs_cs_policy (Resource)

Manages a content switching policy.

## Example Usage

```hcl
data "mcs_cs_action" "route_to_backend" {
  name = "route-to-backend"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_cs_policy" "by_url" {
  name         = "route-by-url"
  action       = data.mcs_cs_action.route_to_backend.id
  expression   = "HTTP.REQ.URL.PATH.STARTSWITH(\"/api\")"
  customer     = data.mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

## Attributes

| Attribute      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `name`        | String | **Yes**  | Policy name. |
| `action`      | String | No       | UUID of the CS action to invoke. |
| `expression`  | String | No       | Policy expression. Computed by the server when omitted. |
| `customer`    | String | No       | Customer identifier. Computed by the server when omitted. |
| `application` | String | No       | UUID of the application. |
| `loadbalancer` | String | No      | UUID of the load balancer. Computed by the server when omitted. |

**Read-only attributes:** `id` (String).

## Import

```shell
terraform import mcs_cs_policy.by_url <cspolicy-uuid>
```
