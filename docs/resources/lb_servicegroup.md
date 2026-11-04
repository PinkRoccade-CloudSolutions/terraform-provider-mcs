---
page_title: "mcs_lb_servicegroup Resource - mcs"
subcategory: "Load Balancing"
description: |-
  Manages a load balancer service group.
---

# mcs_lb_servicegroup (Resource)

Manages a load balancer service group.

> **Note:** The API requires service groups to be created before members can be assigned. When `members` is specified, the provider first creates the service group and then sets its members in a separate update call. This is handled transparently.

## Example Usage

```hcl
data "mcs_lb_servicegroup_member" "web1" {
  id = "11111111-1111-4111-8111-111111111111"
}

data "mcs_lb_servicegroup_member" "web2" {
  id = "22222222-2222-4222-8222-222222222222"
}

data "mcs_lb_monitor" "http_monitor" {
  name = "http-health-check"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_lb_servicegroup" "web_backend" {
  name          = "web-backend-sg"
  type          = "HTTP"
  state         = "enable"
  members       = [data.mcs_lb_servicegroup_member.web1.id, data.mcs_lb_servicegroup_member.web2.id]
  monitors      = [data.mcs_lb_monitor.http_monitor.id]
  healthmonitor = "YES"
  cip           = "ENABLED"
  cipheader     = "X-Forwarded-For"
  customer      = data.mcs_customer.example.id
  loadbalancer  = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

## Attributes

| Attribute       | Type         | Required | Description |
|----------------|-------------|----------|-------------|
| `name`         | String       | **Yes**  | Service group name. |
| `type`         | String       | **Yes**  | Service type: `HTTP`, `SSL`, `ssl_bridge`, `tcp` or `udp`. |
| `state`        | String       | No       | State: `enable` or `disable`. Computed by the server when omitted. |
| `members`      | List(String) | No       | List of member IDs. Set after creation via a separate API call. Removing the attribute clears the members. |
| `monitors`     | List(String) | No       | List of monitor IDs bound to the service group. Removing the attribute clears the monitors. |
| `healthmonitor` | String      | No       | Health monitoring: `YES` or `NO`. Computed by the server when omitted. |
| `client_certificate` | String | No       | UUID of the certificate presented to backend servers that request client authentication. |
| `cip`          | String       | No       | Insert the client IP in a request header: `ENABLED` or `DISABLED`. Computed by the server when omitted. |
| `cipheader`    | String       | No       | Header name for the client IP. The server defaults it to `X-Forwarded-For`. |
| `customer`     | String       | No       | Customer identifier. Computed by the server when omitted. |
| `loadbalancer` | String       | No       | UUID of the load balancer. Computed by the server when omitted. |

**Read-only attributes:** `id` (String).

## Import

```shell
terraform import mcs_lb_servicegroup.web_backend <servicegroup-uuid>
```
