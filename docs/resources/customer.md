---
page_title: "mcs_customer Resource - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Manages a customer within a tenant.
---

# mcs_customer (Resource)

Manages a customer within a tenant.

## Example Usage

```hcl
data "mcs_contact" "admin" {
  name = "Example Corp"
}

resource "mcs_customer" "example" {
  name           = "Example Customer"
  tech_contacts  = [data.mcs_contact.admin.id]
  admin_contacts = [data.mcs_contact.admin.id]
}
```

## Attributes

| Attribute        | Type         | Required | Description |
|-----------------|-------------|----------|-------------|
| `name`          | String       | **Yes**  | Customer name (max 255 characters). |
| `slug`          | String       | No       | Short DNS label (lowercase letters, digits and `-`, max 63), unique within the tenant. Prefixes the customer's Kubernetes cluster names and names its folder in the cluster repository. Set by administrators. Computed from the API when omitted. |
| `contractid`    | String       | No       | Contract identifier (max 255 characters). Computed from the API when omitted. |
| `sdm`           | Number       | No       | Service Delivery Manager user ID. Null when the API has none set. Computed from the API when omitted. |
| `tech_contacts` | List(Number) | No       | List of technical contact IDs. Computed from the API when omitted; set `[]` to clear. |
| `admin_contacts` | List(Number) | No      | List of administrative contact IDs. Computed from the API when omitted; set `[]` to clear. |

**Read-only attributes:** `id` (String) — The customer ID.

## Import

```shell
terraform import mcs_customer.example <customer-id>
```
