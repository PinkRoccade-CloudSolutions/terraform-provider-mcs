---
page_title: "mcs_certificate Resource - mcs"
subcategory: "Load Balancing"
description: |-
  Manages a load balancer SSL certificate.
---

# mcs_certificate (Resource)

Manages a load balancer SSL certificate.

## Example Usage

```hcl
resource "mcs_certificate" "web_cert" {
  name         = "web-ssl-cert"
  ca           = false
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
```

## Attributes

| Attribute          | Type   | Required | Description |
|-------------------|--------|----------|-------------|
| `name`            | String | **Yes**  | Certificate name. |
| `loadbalancer`    | String | **Yes**  | UUID of the load balancer. |
| `ca`              | Bool   | No       | Whether this is a CA certificate. Computed by the server when omitted. |
| `valid_to_timestamp` | String | No    | Certificate expiry timestamp. Computed by the server when omitted. |
| `customer`        | String | No       | Customer identifier. Computed by the server when omitted. |

**Read-only attributes:** `id` (String), `protected` (Bool, whether the certificate is protected from changes).

## Import

```shell
terraform import mcs_certificate.web_cert <certificate-uuid>
```
