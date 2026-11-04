---
page_title: "mcs_certificate Data Source - mcs"
subcategory: "Load Balancing"
description: |-
  Look up load balancer certificates. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_certificate (Data Source)

Look up load balancer certificates. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_certificate" "web_cert" {
  name = "web-ssl-cert"
}
```

## Attributes

| Attribute            | Type   | Mode     | Description |
|---------------------|--------|----------|-------------|
| `name`              | String | Optional | Exact certificate name. |
| `id`                | String | Optional/Computed | Certificate UUID. |
| `ca`                | Bool   | Computed | Whether this is a CA certificate. |
| `valid_to_timestamp` | String | Computed | Certificate expiry timestamp. |
| `loadbalancer`      | String | Computed | UUID of the load balancer. |
| `customer`          | String | Computed | Customer identifier. |
| `protected`         | Bool   | Computed | Whether the certificate is protected from changes. |
| `certificates`      | List   | Computed | All certificates (populated when neither `name` nor `id` is set). |

**Nested `certificates` attributes:** `id`, `name`, `valid_to_timestamp`, `loadbalancer`, `customer` (String); `ca`, `protected` (Bool) — all Computed.
