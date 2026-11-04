---
page_title: "mcs_quota Data Source - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Look up quota profiles (a combination of a domain quota and a loadbalancing quota) by `id` or exact `name`, or list all of them.
---

# mcs_quota (Data Source)

Look up quota profiles. A quota profile combines a domain quota profile and a loadbalancing quota profile. Provide `name` (exact match) or `id` for a single profile, or omit both to list all profiles in `quotas`.

## Example Usage

```hcl
data "mcs_quota" "standard" {
  name = "standard"
}

# Resolve the domain quota profile referenced by the quota profile.
data "mcs_domain_quota" "standard" {
  id = data.mcs_quota.standard.domain_quota
}
```

## Attributes

| Attribute             | Type   | Mode              | Description |
|-----------------------|--------|-------------------|-------------|
| `name`                | String | Optional/Computed | Exact quota name to look up. |
| `id`                  | String | Optional/Computed | ID of a specific quota to look up. |
| `description`         | String | Computed          | Description of the quota profile. |
| `domain_quota`        | String | Computed          | ID of the domain quota profile. |
| `loadbalancing_quota` | String | Computed          | ID of the loadbalancing quota profile. |
| `quotas`              | List   | Computed          | All quota profiles (populated when neither `name` nor `id` is set). |

**Nested `quotas` attributes:** `id`, `name`, `description`, `domain_quota`, `loadbalancing_quota` (String) — all Computed.
