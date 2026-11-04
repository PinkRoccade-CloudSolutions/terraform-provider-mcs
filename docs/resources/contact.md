---
page_title: "mcs_contact Resource - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Manages a contact record for a tenant.
---

# mcs_contact (Resource)

Manages a contact record for a tenant.

## Example Usage

```hcl
resource "mcs_contact" "admin" {
  company   = "Example Corp"
  firstname = "John"
  lastname  = "Doe"
  email     = "john.doe@example.com"
  phone     = "+31 6 12345678"
}
```

## Attributes

| Attribute   | Type   | Required | Description |
|------------|--------|----------|-------------|
| `company`  | String | **Yes**  | Company name (max 255 characters). |
| `firstname` | String | No      | First name. |
| `lastname` | String | No       | Last name. |
| `email`    | String | No       | Email address. |
| `phone`    | String | No       | Phone number. |
| `address`  | String | No       | Address. |

All optional fields are at most 255 characters. When omitted, the value stored by the API (usually `""`) is kept in state; removing a field from the configuration does not clear it — set it to `""` instead.

**Read-only attributes:** `id` (String) — The contact ID.

## Import

```shell
terraform import mcs_contact.admin <contact-id>
```
