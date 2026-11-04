---
page_title: "mcs_domain_quota Data Source - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Look up domain quota profiles by `id` or exact `name`, or list all of them.
---

# mcs_domain_quota (Data Source)

Look up domain quota profiles, which limit the number of firewall objects in a domain. Provide `name` (exact match) or `id` for a single profile, or omit both to list all profiles in `domain_quotas`.

## Example Usage

```hcl
data "mcs_domain_quota" "standard" {
  name = "standard"
}

output "max_firewall_rules" {
  value = data.mcs_domain_quota.standard.max_rules
}
```

### List all domain quota profiles

```hcl
data "mcs_domain_quota" "all" {}

output "domain_quota_names" {
  value = [for q in data.mcs_domain_quota.all.domain_quotas : q.name]
}
```

## Attributes

| Attribute           | Type   | Mode              | Description |
|---------------------|--------|-------------------|-------------|
| `name`              | String | Optional/Computed | Exact domain quota name to look up. |
| `id`                | String | Optional/Computed | ID of a specific domain quota to look up. |
| `description`       | String | Computed          | Description of the quota profile. |
| `max_rules`         | Number | Computed          | Max number of firewall rules in a domain. |
| `max_objects`       | Number | Computed          | Max number of firewall objects in a domain. |
| `max_objectgroups`  | Number | Computed          | Max number of firewall object groups in a domain. |
| `max_services`      | Number | Computed          | Max number of firewall services in a domain. |
| `max_servicegroups` | Number | Computed          | Max number of firewall service groups in a domain. |
| `max_networks`      | Number | Computed          | Max number of networks (VLANs) in a domain. |
| `domain_quotas`     | List   | Computed          | All domain quota profiles (populated when neither `name` nor `id` is set). |

**Nested `domain_quotas` attributes:** `id`, `name`, `description` (String); `max_rules`, `max_objects`, `max_objectgroups`, `max_services`, `max_servicegroups`, `max_networks` (Number) — all Computed.
