---
page_title: "mcs_csv_server Resource - mcs"
subcategory: "Load Balancing"
description: |-
  Manages a content switching virtual server (CS vServer).
---

# mcs_csv_server (Resource)

Manages a content switching virtual server (CS vServer).

## Example Usage

```hcl
data "mcs_public_ip_address" "web" {
  ip_address = "203.0.113.10"
}

data "mcs_cs_policy" "routing" {
  name = "route-by-url"
}

data "mcs_certificate" "web_cert" {
  name = "web-cert"
}

data "mcs_certificate" "client_ca" {
  name = "client-ca"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_csv_server" "web_frontend" {
  name           = "web-csvserver"
  ufname         = "web-frontend"
  type           = "ssl"
  port           = 443
  ipaddress      = data.mcs_public_ip_address.web.id
  policies       = [data.mcs_cs_policy.routing.id]
  certificate    = [data.mcs_certificate.web_cert.id]
  ca_certificate = [data.mcs_certificate.client_ca.id]
  clientauth     = true
  clientcert     = "Mandatory"
  customer       = data.mcs_customer.example.id
  loadbalancer   = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

## Attributes

| Attribute      | Type         | Required | Description |
|---------------|-------------|----------|-------------|
| `name`        | String       | **Yes**  | CS vServer name. |
| `ufname`      | String       | **Yes**  | User-friendly name. |
| `type`        | String       | **Yes**  | Protocol type: `http` or `ssl`. |
| `ipaddress`   | String       | No       | UUID of the associated PublicIPAddress. |
| `port`        | Number       | No       | Listening port. Defaults to `443`. |
| `policies`    | List(String) | No       | List of CS policy IDs. Removing the attribute clears the policies. |
| `certificate` | List(String) | No       | List of SSL certificate IDs. Removing the attribute clears the certificates. |
| `ca_certificate` | List(String) | No    | List of CA certificate IDs used to verify client certificates. Removing the attribute clears them. |
| `clientauth`  | Bool         | No       | Enable client certificate authentication. Computed by the server when omitted. |
| `clientcert`  | String       | No       | Client certificate requirement: `Mandatory` or `Optional`. Computed by the server when omitted. |
| `customer`    | String       | No       | Customer identifier. Computed by the server when omitted. |
| `loadbalancer` | String      | No       | UUID of the load balancer. Computed by the server when omitted. |

**Read-only attributes:** `id` (String).

## Import

```shell
terraform import mcs_csv_server.web_frontend <csvserver-uuid>
```
