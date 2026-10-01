<!--
Pending docs for mcs_application. Merge into docs/index.md:
- TOC, under "Data Sources" (after mcs_cs_policy):
    - [mcs_application](#mcs_application-data-source)
- TOC, under "Resources" > "Load Balancing" (before mcs_cs_policy):
    - [mcs_application](#mcs_application)
- Data source section: in the Data Sources part (before "### mcs_rewrite_action (Data Source)").
- Resource section: in "### Load Balancing" (before "#### mcs_cs_policy").
-->

### mcs_application (Data Source)

Look up applications in the tenant's application catalogue. Provide `name` or `id` for a single match, or omit both to list all. The API offers no list filters, so `name` is matched exactly client-side.

#### Example

```hcl
data "mcs_application" "webshop" {
  name = "webshop"
}

resource "mcs_cs_policy" "webshop" {
  # ...
  application = data.mcs_application.webshop.id
}
```

#### Attributes

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

---

#### mcs_application

Manages an application in the tenant's application catalogue. Applications are referenced by `mcs_cs_policy.application`.

##### Example

```hcl
resource "mcs_application" "webshop" {
  name             = "webshop"
  pentested        = true
  pentest_type     = "graybox"
  pentest_date     = "2026-03-15"
  pentest_findings = 4
}
```

##### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | **Yes** | Name of the application (max 255 characters). |
| `pentested` | Bool | No | Whether the application has been pentested. Server default when unset. |
| `pentest_type` | String | No | Type of pentest: `blackbox`, `graybox`, `whitebox` or `unknown`. Server default when unset. |
| `pentest_date` | String | No | Date of the pentest (`YYYY-MM-DD`). Removing it clears the value. |
| `pentest_findings` | Number | No | Number of findings in the pentest. Removing it clears the value. |

**Read-only attributes:** `id` (String).

##### Import

```shell
terraform import mcs_application.webshop <application-uuid>
```

---
