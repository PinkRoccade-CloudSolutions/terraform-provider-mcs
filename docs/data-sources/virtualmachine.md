---
page_title: "mcs_virtualmachine Data Source - mcs"
subcategory: "Compute"
description: |-
  Look up virtual machines. Provide `name` or `id` for a single VM, or omit both to list all VMs. Returns detailed information including disks and network interfaces.
---

# mcs_virtualmachine (Data Source)

Look up virtual machines. Provide `name` or `id` for a single VM, or omit both to list all VMs. Returns detailed information including disks and network interfaces.

## Example Usage — Single lookup

```hcl
data "mcs_virtualmachine" "webserver" {
  name = "web-server-01"
}

output "vm_cpu" {
  value = data.mcs_virtualmachine.webserver.cpu
}

output "vm_primary_ip" {
  value = data.mcs_virtualmachine.webserver.interfaces[0].ipaddress
}
```

## Example Usage — List all

```hcl
data "mcs_virtualmachine" "all" {}

output "vm_names" {
  value = [for vm in data.mcs_virtualmachine.all.virtual_machines : vm.name]
}
```

## Example Usage — VMs in a domain

```hcl
data "mcs_domain" "production" {
  name = "production"
}

data "mcs_virtualmachine" "in_domain" {
  domain_id = data.mcs_domain.production.id
}
```

## Attributes

| Attribute          | Type   | Mode     | Description |
|-------------------|--------|----------|-------------|
| `name`            | String | Optional | Exact VM name to look up (sent as `name__icontains`, matched exactly client-side). |
| `id`              | String | Optional/Computed | VM UUID. |
| `domain_id`       | Number | Optional | Only consider VMs with an interface in a network of this domain (`interfaces__network__domain__id`). Combines with `name`. |
| `cpu`             | Number | Computed | Number of vCPUs (set when a single VM is matched). |
| `memory`          | Number | Computed | Memory in MB (set when a single VM is matched). |
| `os`              | String | Computed | Operating system (set when a single VM is matched). |
| `state`           | String | Computed | Power state (set when a single VM is matched). |
| `disks`           | List   | Computed | Disks attached to the matched VM. |
| `interfaces`      | List   | Computed | Network interfaces of the matched VM. |
| `virtual_machines` | List  | Computed | All VMs (populated when neither `name` nor `id` is set). Each item has `id`, `name`, `cpu`, `memory`, `os`, `state`, `disks`, `interfaces`. |

**Nested `disks` attributes:**

| Attribute | Type   | Description |
|----------|--------|-------------|
| `id`     | String | Disk ID. |
| `name`   | String | Disk name. |
| `size`   | Number | Size in GB. |
| `path`   | String | Disk path. |
| `type`   | String | Disk type. |

**Nested `interfaces` attributes:**

| Attribute     | Type   | Description |
|--------------|--------|-------------|
| `id`         | String | Interface ID. |
| `name`       | String | Interface name. |
| `ipaddress`  | String | IPv4 address. |
| `ipv6address` | String | IPv6 address. |
| `network`    | String | Associated network UUID (null when not connected). |
| `mac_address` | String | MAC address. |
| `vm_name`    | String | Name of the VM the interface belongs to. |
