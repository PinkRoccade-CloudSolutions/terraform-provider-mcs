---
page_title: "mcs_loadbalancing_quota Data Source - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Look up loadbalancing quota profiles by `id` or exact `name`, or list all of them.
---

# mcs_loadbalancing_quota (Data Source)

Look up loadbalancing quota profiles. Provide `name` (exact match) or `id` for a single profile, or omit both to list all profiles in `loadbalancing_quotas`.

## Example Usage

```hcl
data "mcs_loadbalancing_quota" "standard" {
  name = "standard"
}

output "max_policies_per_csvserver" {
  value = data.mcs_loadbalancing_quota.standard.max_policies_in_csvserver
}
```

## Attributes

| Attribute                   | Type   | Mode              | Description |
|-----------------------------|--------|-------------------|-------------|
| `name`                      | String | Optional/Computed | Exact loadbalancing quota name to look up. |
| `id`                        | String | Optional/Computed | ID of a specific loadbalancing quota to look up. |
| `description`               | String | Computed          | Description of the quota profile. |
| `max_policies_in_csvserver` | Number | Computed          | Max number of policies linked to a content switching virtual server. |
| `loadbalancing_quotas`      | List   | Computed          | All loadbalancing quota profiles (populated when neither `name` nor `id` is set). |

**Nested `loadbalancing_quotas` attributes:** `id`, `name`, `description` (String); `max_policies_in_csvserver` (Number) — all Computed.
