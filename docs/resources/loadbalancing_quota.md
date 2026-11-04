---
page_title: "mcs_loadbalancing_quota Resource - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Manages a loadbalancing quota profile. Quota profiles are global (not tenant-scoped).
---

# mcs_loadbalancing_quota (Resource)

Manages a loadbalancing quota profile. Quota profiles are global (not tenant-scoped).

## Example Usage

```hcl
resource "mcs_loadbalancing_quota" "standard" {
  name                      = "standard"
  max_policies_in_csvserver = 40
}
```

## Attributes

| Attribute                   | Type   | Required | Description |
|-----------------------------|--------|----------|-------------|
| `name`                      | String | **Yes**  | Profile name. |
| `description`               | String | No       | Description. |
| `max_policies_in_csvserver` | Number | No       | Max number of policies linked to a content switching virtual server. API default when omitted. |

**Read-only attributes:** `id` (String).


## Import

```shell
terraform import mcs_loadbalancing_quota.standard <loadbalancing-quota-uuid>
```
