---
page_title: "mcs_contact Data Source - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Look up tenant contacts. Provide `name` (matches company) or `id` for a single match, or omit both to list all.
---

# mcs_contact (Data Source)

Look up tenant contacts. Provide `name` (matches company) or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_contact" "admin" {
  name = "Example Corp"
}
```

## Attributes

| Attribute   | Type   | Mode     | Description |
|------------|--------|----------|-------------|
| `name`     | String | Optional | Exact company name to look up (sent as `company__icontains`, matched exactly client-side). |
| `id`       | String | Optional/Computed | Contact ID. |
| `company`  | String | Computed | Company name. |
| `firstname` | String | Computed | First name. |
| `lastname` | String | Computed | Last name. |
| `email`    | String | Computed | Email address. |
| `phone`    | String | Computed | Phone number. |
| `address`  | String | Computed | Address. |
| `contacts` | List   | Computed | All contacts (populated when neither `name` nor `id` is set). |
