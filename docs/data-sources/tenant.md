---
page_title: "mcs_tenant Data Source - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Retrieves information about the current tenant.
---

# mcs_tenant (Data Source)

Retrieves information about the current tenant.

## Example Usage

```hcl
data "mcs_tenant" "current" {}

output "tenant_name" {
  value = data.mcs_tenant.current.name
}
```

## Attributes

All attributes are computed (read-only):

| Attribute    | Type   | Description |
|-------------|--------|-------------|
| `id`        | String | Tenant identifier. |
| `name`      | String | Tenant name. |
| `description` | String | Tenant description. |
| `raw_json`  | String | Full JSON response body (useful for debugging). |
