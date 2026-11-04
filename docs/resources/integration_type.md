---
page_title: "mcs_integration_type Resource - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Manages an integration type within the tenant.
---

# mcs_integration_type (Resource)

Manages an integration type within the tenant.

## Example Usage

```hcl
resource "mcs_integration_type" "monitoring" {
  name = "monitoring"
}
```

## Attributes

| Attribute | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `name`    | String | **Yes**  | Integration type name. |

**Read-only attributes:** `id` (Number).


## Import

The import ID is the numeric ID of the integration type.

```shell
terraform import mcs_integration_type.monitoring <integration-type-id>
```
