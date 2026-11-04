---
page_title: "mcs_domain_quota Resource - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Manages a domain quota profile that limits the number of firewall objects in a domain. Quota profiles are global (not tenant-scoped). Limits that are omitted get the API default.
---

# mcs_domain_quota (Resource)

Manages a domain quota profile that limits the number of firewall objects in a domain. Quota profiles are global (not tenant-scoped). Limits that are omitted get the API default.

## Example Usage

```hcl
resource "mcs_domain_quota" "standard" {
  name         = "standard"
  description  = "Standard tier"
  max_rules    = 750
  max_objects  = 1500
  max_networks = 20
}
```

## Attributes

| Attribute           | Type   | Required | Description |
|---------------------|--------|----------|-------------|
| `name`              | String | **Yes**  | Profile name. |
| `description`       | String | No       | Description. |
| `max_rules`         | Number | No       | Max number of firewall rules in a domain. |
| `max_objects`       | Number | No       | Max number of firewall objects in a domain. |
| `max_objectgroups`  | Number | No       | Max number of firewall object groups in a domain. |
| `max_services`      | Number | No       | Max number of firewall services in a domain. |
| `max_servicegroups` | Number | No       | Max number of firewall service groups in a domain. |
| `max_networks`      | Number | No       | Max number of networks (VLANs) in a domain. |

**Read-only attributes:** `id` (String).


## Import

```shell
terraform import mcs_domain_quota.standard <domain-quota-uuid>
```
