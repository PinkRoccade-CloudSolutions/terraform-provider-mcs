---
page_title: "mcs_application Resource - mcs"
subcategory: "Load Balancing"
description: |-
  Manages an application in the tenant's application catalogue. Applications are referenced by `mcs_cs_policy.application`.
---

# mcs_application (Resource)

Manages an application in the tenant's application catalogue. Applications are referenced by `mcs_cs_policy.application`.

## Example Usage

```hcl
resource "mcs_application" "webshop" {
  name             = "webshop"
  pentested        = true
  pentest_type     = "graybox"
  pentest_date     = "2026-03-15"
  pentest_findings = 4
}
```

## Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | **Yes** | Name of the application (max 255 characters). |
| `pentested` | Bool | No | Whether the application has been pentested. Server default when unset. |
| `pentest_type` | String | No | Type of pentest: `blackbox`, `graybox`, `whitebox` or `unknown`. Server default when unset. |
| `pentest_date` | String | No | Date of the pentest (`YYYY-MM-DD`). Removing it clears the value. |
| `pentest_findings` | Number | No | Number of findings in the pentest. Removing it clears the value. |

**Read-only attributes:** `id` (String).

## Import

```shell
terraform import mcs_application.webshop <application-uuid>
```
