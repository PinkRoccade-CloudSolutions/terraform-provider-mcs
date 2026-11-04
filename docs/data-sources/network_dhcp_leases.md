---
page_title: "mcs_network_dhcp_leases Data Source - mcs"
subcategory: "Networking"
description: |-
  Lists the static DHCP leases of a network, including leases the device sync found on the firewall.
---

# mcs_network_dhcp_leases (Data Source)

Lists the static DHCP leases of a network, including leases the device sync found on the firewall. Use it to find the UUID of a lease to import into `mcs_network_dhcp_lease`.

## Example Usage

```hcl
data "mcs_network" "office" {
  name = "NET-OFFICE"
}

data "mcs_network_dhcp_leases" "office" {
  network = data.mcs_network.office.id
}

output "reserved_ips" {
  value = { for l in data.mcs_network_dhcp_leases.office.leases : l.mac => l.ip }
}
```

## Attributes

| Attribute | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `network` | String | **Yes**  | UUID of the network whose leases to list. |
| `ip`      | String | No       | Only list the lease for this IPv4 address. |
| `mac`     | String | No       | Only list the lease for this MAC address, in any common notation. |

**Read-only attributes:**

| Attribute | Type | Description |
|-----------|------|-------------|
| `leases`  | List | The matching leases. Each has `id`, `network`, `ip`, `mac` (lower case, colon separated), `description`, `created_at` and `updated_at`. |
