---
page_title: "mcs_integration_type Data Source - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Look up integration types. Provide `name` (exact match) or `id` for a single type, or omit both to list all of them.
---

# mcs_integration_type (Data Source)

Look up integration types. Provide `name` (exact match) or `id` for a single type, or omit both to list all of them.

## Example Usage

```hcl
data "mcs_integration_type" "monitoring" {
  name = "monitoring"
}
```

## Attributes

| Attribute           | Type   | Mode              | Description |
|---------------------|--------|-------------------|-------------|
| `id`                | Number | Optional/Computed | Integration type ID. |
| `name`              | String | Optional/Computed | Exact integration type name. |
| `integration_types` | List   | Computed          | All integration types (populated when neither `name` nor `id` is set). |
