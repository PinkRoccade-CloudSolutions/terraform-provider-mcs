---
page_title: "mcs_quota Resource - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Manages a quota profile that combines a domain quota and a loadbalancing quota. Quota profiles are global (not tenant-scoped).
---

# mcs_quota (Resource)

Manages a quota profile that combines a domain quota and a loadbalancing quota. Quota profiles are global (not tenant-scoped).

## Example Usage

```hcl
data "mcs_domain_quota" "standard" {
  name = "standard"
}

data "mcs_loadbalancing_quota" "standard" {
  name = "standard"
}

resource "mcs_quota" "standard" {
  name                = "standard"
  domain_quota        = data.mcs_domain_quota.standard.id
  loadbalancing_quota = data.mcs_loadbalancing_quota.standard.id
}
```

## Attributes

| Attribute             | Type   | Required | Description |
|-----------------------|--------|----------|-------------|
| `name`                | String | **Yes**  | Profile name. |
| `description`         | String | No       | Description. |
| `domain_quota`        | String | No       | ID of the domain quota profile. |
| `loadbalancing_quota` | String | No       | ID of the loadbalancing quota profile. |

**Read-only attributes:** `id` (String).


## Import

```shell
terraform import mcs_quota.standard <quota-uuid>
```
