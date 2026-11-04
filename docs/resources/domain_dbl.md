---
page_title: "mcs_domain_dbl Resource - mcs"
subcategory: "Deny/Block Lists"
description: |-
  Manages a domain name entry in the deny/block list.
---

# mcs_domain_dbl (Resource)

Manages a domain name entry in the deny/block list.

## Example Usage

```hcl
resource "mcs_domain_dbl" "blocked_domain" {
  domainname = "malicious-site.example"
  source     = "manual"
  persistent = true
  labels     = ["malware"]
}
```

## Attributes

| Attribute    | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `domainname` | String | **Yes** | Domain name to block (5–255 characters). |
| `source`    | String | No       | Source of the block entry (5–255 characters). Computed from the API when omitted. |
| `persistent` | Bool  | No       | Keep the entry instead of expiring it after the default time. Computed from the API when omitted. |
| `labels`    | List(String) | No | Names of DBL labels to attach (sent as nested `{"name": ...}` objects). Computed from the API when omitted; set `[]` to clear. |

**Read-only attributes:**

| Attribute    | Type   | Description |
|-------------|--------|-------------|
| `id`        | String | Entry ID. |
| `timestamp` | String | Creation timestamp. |
| `occurrence` | Number | Number of occurrences. |

## Import

```shell
terraform import mcs_domain_dbl.blocked_domain <domain-dbl-id>
```
