---
page_title: "mcs_secureingress_firewall Resource - mcs"
subcategory: "Secure Ingress"
description: |-
  Manages a secure ingress (XDP) firewall filter. Referenced by `mcs_ingress_cluster.firewall`.
---

# mcs_secureingress_firewall (Resource)

Manages a secure ingress (XDP) firewall filter. Referenced by `mcs_ingress_cluster.firewall`.

The API stores `filter_json` as a JSON-encoded string. Write it with `jsonencode()`; differences in whitespace or key order between your configuration and the API's response do not cause a diff.

## Example Usage

```hcl
data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_secureingress_firewall" "edge" {
  name        = "edge-filter"
  description = "Blocks unwanted traffic"
  customer    = data.mcs_customer.example.id
  filter_json = jsonencode({
    allow = ["10.0.0.0/8"]
  })
}
```

## Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | **Yes** | Name of the firewall (max 200 characters). |
| `filter_json` | String (JSON) | **Yes** | Firewall filter definition as a JSON document. |
| `customer` | String | **Yes** | Customer identifier. Changing this forces a new resource. |
| `description` | String | No | Description of the firewall (max 200 characters). |

**Read-only attributes:** `id`, `created_at_timestamp`, `updated_at_timestamp` (String), `created_by_user`, `updated_by_user` (Number).

## Import

```shell
terraform import mcs_secureingress_firewall.edge <firewall-uuid>
```
