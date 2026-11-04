---
page_title: "mcs_application Data Source - mcs"
subcategory: "Load Balancing"
description: |-
  Look up applications in the tenant's application catalogue. Provide `name` or `id` for a single match, or omit both to list all. The API offers no list filters, so `name` is matched exactly client-side.
---

# mcs_application (Data Source)

Look up applications in the tenant's application catalogue. Provide `name` or `id` for a single match, or omit both to list all. The API offers no list filters, so `name` is matched exactly client-side.

## Example Usage

```hcl
data "mcs_application" "webshop" {
  name = "webshop"
}

data "mcs_cs_action" "webshop" {
  name = "route-to-webshop"
}

resource "mcs_cs_policy" "webshop" {
  name        = "route-webshop"
  action      = data.mcs_cs_action.webshop.id
  expression  = "HTTP.REQ.HOSTNAME.EQ(\"shop.example.com\")"
  application = data.mcs_application.webshop.id
}
```

## Attributes

| Attribute | Type | Mode | Description |
|-----------|------|------|-------------|
| `name` | String | Optional/Computed | Exact application name. |
| `id` | String | Optional/Computed | Application UUID. |
| `pentested` | Bool | Computed | Whether the application has been pentested. |
| `pentest_type` | String | Computed | Type of pentest: `blackbox`, `graybox`, `whitebox` or `unknown`. |
| `pentest_date` | String | Computed | Date of the pentest (`YYYY-MM-DD`). |
| `pentest_findings` | Number | Computed | Number of findings in the pentest. |
| `applications` | List | Computed | All applications (populated when neither `name` nor `id` is set). |

**Nested `applications` attributes:** `id`, `name`, `pentested` (Bool), `pentest_type`, `pentest_date`, `pentest_findings` (Number) — all Computed.
