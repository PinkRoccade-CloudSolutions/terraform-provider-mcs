---
page_title: "mcs_lb_servicegroup_member Resource - mcs"
subcategory: "Load Balancing"
description: |-
  Manages a member of a load balancer service group.
---

# mcs_lb_servicegroup_member (Resource)

Manages a member of a load balancer service group.

## Example Usage

```hcl
data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_lb_servicegroup_member" "web1" {
  address      = "10.0.1.10"
  port         = 8080
  servername   = "web-server-01"
  weight       = 100
  state        = "UP"
  customer     = data.mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

## Attributes

| Attribute      | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `address`     | String | **Yes**  | IP address of the backend server. |
| `servername`  | String | **Yes**  | Server name. |
| `port`        | Number | No       | Port number. Defaults to `0`. |
| `weight`      | Number | No       | Load balancing weight. Computed by the server when omitted. |
| `state`       | String | No       | Member state: `UP` or `DOWN`. Computed by the server when omitted. |
| `customer`    | String | No       | Customer identifier. Computed by the server when omitted. |
| `loadbalancer` | String | No      | UUID of the load balancer. Computed by the server when omitted. |

**Read-only attributes:** `id` (String).

## Import

```shell
terraform import mcs_lb_servicegroup_member.web1 <servicegroup-member-uuid>
```
