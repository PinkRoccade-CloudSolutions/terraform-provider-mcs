<!--
Pending docs for the new mcs_dns_domain resource.
TOC entry (under "  - [DNS](#dns)", before mcs_dns_entry):
    - [mcs_dns_domain](#mcs_dns_domain)
Section goes under "### DNS", before "#### mcs_dns_entry".
Note: the data source heading "### mcs_dns_domain (Data Source)" has anchor #mcs_dns_domain-data-source,
so #mcs_dns_domain resolves to this resource section.
-->

#### mcs_dns_domain

Manages the settings of an **existing** MCS DNS domain (zone).

> **Important:** DNS zones cannot be created or deleted through the MCS API. They are synchronised from the DNS provider integration.
>
> - **Create adopts an existing zone.** The provider looks up the zone whose name exactly matches `name` and applies the configured settings to it (`PUT /api/dns/domains/{uuid}/`). If no zone with that name exists, the apply fails with "DNS domain not found" — make sure the zone exists at the DNS provider and has been synchronised into MCS first.
> - **Destroy does not delete the zone.** Removing the resource (or running `terraform destroy`) only removes it from Terraform state and emits a warning. The zone and its records remain in MCS and at the DNS provider.
> - Optional attributes that are not configured keep their current value in MCS; only configured attributes are sent.

##### Example

```hcl
resource "mcs_dns_domain" "example" {
  name      = "example.com"
  comment   = "Managed by Terraform"
  type      = "external"
  zone_type = "forward"
  customer  = mcs_customer.example.id
}

resource "mcs_dns_entry" "www" {
  domain_uuid = mcs_dns_domain.example.id
  name        = "www"
  type        = "A"
  content     = "192.0.2.1"
  expire      = 300
}
```

##### Attributes

| Attribute   | Type   | Required | Description |
|------------|--------|----------|-------------|
| `name`      | String | **Yes**  | Full zone name as known by the DNS provider (e.g. `example.com`). Must match an existing zone exactly. Changing this forces a new resource (the old zone is released from state, the new one adopted). |
| `comment`   | String | No       | Comment for the domain (max 255 characters). Computed from MCS when omitted. |
| `enddate`   | String | No       | End date for the domain (`YYYY-MM-DD`), if known. Computed from MCS when omitted. |
| `customer`  | String | No       | Customer associated with the domain. Computed from MCS when omitted. |
| `type`      | String | No       | Domain use type: `external` or `internal`. Computed from MCS when omitted. |
| `zone_type` | String | No       | Zone type: `forward` or `reverse`. Computed from MCS when omitted. |

**Read-only attributes:**

| Attribute       | Type   | Description |
|----------------|--------|-------------|
| `id`            | String | UUID of the DNS domain. |
| `provider_id`   | Number | ID of the DNS provider integration. |
| `provider_name` | String | Name of the DNS provider integration. |

##### Import

Existing zones can also be imported by UUID:

```shell
terraform import mcs_dns_domain.example 0b7c0f9e-1234-4cde-9abc-0123456789ab
```

---
