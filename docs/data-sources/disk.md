---
page_title: "mcs_disk Data Source - mcs"
subcategory: "Compute"
description: |-
  Look up virtual machine disks. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_disk (Data Source)

Look up virtual machine disks. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_disk" "primary" {
  name = "sda"
}

output "disk_size" {
  value = data.mcs_disk.primary.size
}
```

## Attributes

| Attribute | Type   | Mode     | Description |
|----------|--------|----------|-------------|
| `name`   | String | Optional | Exact disk name to look up (the API has no name filter, so all disks are listed and matched client-side). |
| `id`     | String | Optional/Computed | Disk UUID. |
| `size`   | Number | Computed | Size in GB. |
| `path`   | String | Computed | Disk path. |
| `type`   | String | Computed | Disk type: `thin` or `thick`. |
| `disks`  | List   | Computed | All disks (populated when neither `name` nor `id` is set). |

**Nested `disks` attributes:** `id`, `name`, `path`, `type` (String); `size` (Number) — all Computed.
