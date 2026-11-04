---
page_title: "mcs_dns_entry Resource - mcs"
subcategory: "DNS"
description: |-
  Manages a DNS entry within an MCS DNS domain. Use the `mcs_dns_domain` data source to look up available domains and their UUIDs.
---

# mcs_dns_entry (Resource)

Manages a DNS entry within an MCS DNS domain. Use the `mcs_dns_domain` data source to look up available domains and their UUIDs.

DNS entries cannot be updated in-place. Changing any attribute will cause Terraform to destroy the existing entry and create a new one.

## Example Usage

```hcl
data "mcs_dns_domain" "main" {
  name = "example.com"
}

resource "mcs_dns_entry" "www" {
  domain_uuid = data.mcs_dns_domain.main.domains[0].uuid
  name        = "www"
  type        = "A"
  content     = "192.0.2.1"
  expire      = 300
}

resource "mcs_dns_entry" "mail" {
  domain_uuid = data.mcs_dns_domain.main.domains[0].uuid
  name        = "mail"
  type        = "CNAME"
  content     = "mail.example.com"
  expire      = 3600
}
```

## Attributes

| Attribute     | Type   | Required | Description |
|--------------|--------|----------|-------------|
| `domain_uuid` | String | **Yes** | UUID of the DNS domain (from `mcs_dns_domain` data source). Changing this forces a new resource. |
| `name`       | String | **Yes**  | DNS record name (e.g. www, mail). Changing this forces a new resource. |
| `type`       | String | **Yes**  | DNS record type (e.g. A, AAAA, CNAME, MX, TXT). Changing this forces a new resource. |
| `content`    | String | **Yes**  | DNS record content (e.g. IP address, hostname). Changing this forces a new resource. |
| `expire`     | Number | **Yes**  | TTL in seconds, between 60 and 604800 (validated at plan time). Changing this forces a new resource. |

**Read-only attributes:**

| Attribute | Type   | Description |
|-----------|--------|-------------|
| `id`      | String | Composite identifier: `domain_uuid/name/type/content`. |

## Import

Import by the composite ID `<domain_uuid>/<name>/<type>/<content>`. The entry is found by an exact match on name, type and content, so give each one exactly as MCS lists it for the zone. Depending on the zone, MCS returns names relative (`www`) or as an FQDN with a trailing dot (`www.example.com.`). Content is the last part and may itself contain `/`.

```shell
terraform import mcs_dns_entry.www 0b7c0f9e-1234-4cde-9abc-0123456789ab/www/A/192.0.2.1
```
