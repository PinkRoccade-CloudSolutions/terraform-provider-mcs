---
page_title: "mcs_lbv_server Resource - mcs"
subcategory: "Load Balancing"
description: |-
  Manages a load balancer virtual server (LB vServer).
---

# mcs_lbv_server (Resource)

Manages a load balancer virtual server (LB vServer).

## Example Usage

```hcl
data "mcs_public_ip_address" "web" {
  ip_address = "203.0.113.10"
}

data "mcs_lb_servicegroup" "web_backend" {
  name = "web-backend-sg"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_lbv_server" "web_lb" {
  name         = "web-lbvserver"
  ipaddress    = data.mcs_public_ip_address.web.id
  port         = 443
  type         = "ssl"
  servicegroup = [data.mcs_lb_servicegroup.web_backend.id]
  customer     = data.mcs_customer.example.id
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

## Attributes

| Attribute      | Type         | Required | Description |
|---------------|-------------|----------|-------------|
| `name`        | String       | **Yes**  | Virtual server name. |
| `servicegroup` | List(String) | **Yes** | List of service group IDs to bind. |
| `ipaddress`   | String       | No       | UUID of the associated PublicIPAddress. Leave empty if used as non routed loadbalancer. |
| `port`        | Number       | No       | Listening port. Defaults to `0` (non routed loadbalancer); a null port returned by the API is shown as `0`. |
| `type`        | String       | No       | Protocol type: `http`, `ssl`, `ssl_bridge`, `tcp` or `udp`. Defaults to `ssl`. |
| `certificate` | List(String) | No       | List of SSL certificate IDs. Removing the attribute clears the certificates. |
| `ca_certificate` | List(String) | No    | List of CA certificate IDs. Removing the attribute clears them. |
| `customer`    | String       | No       | Customer identifier. Computed by the server when omitted. |
| `loadbalancer` | String      | No       | UUID of the load balancer. Computed by the server when omitted. |

**Read-only attributes:** `id` (String).

## Import

```shell
terraform import mcs_lbv_server.web_lb <lbvserver-uuid>
```
