---
page_title: "mcs_virtual_datacenter Data Source - mcs"
subcategory: "Compute"
description: |-
  Look up virtual datacenters. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_virtual_datacenter (Data Source)

Look up virtual datacenters. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_virtual_datacenter" "prod" {
  name = "production-vdc"
}
```

## Attributes

| Attribute             | Type   | Mode     | Description |
|----------------------|--------|----------|-------------|
| `name`               | String | Optional | Exact virtual datacenter name (sent as `name__icontains`, matched exactly client-side). |
| `id`                 | String | Optional/Computed | Virtual datacenter UUID. |
| `cluster`            | String | Computed | UUID of the cluster the VDC deploys VMs on. |
| `virtual_datacenters` | List  | Computed | All virtual datacenters (populated when neither `name` nor `id` is set). |

**Nested `virtual_datacenters` attributes:** `id`, `name`, `cluster` — all String, Computed.
