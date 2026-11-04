---
page_title: "mcs_network Data Source - mcs"
subcategory: "Networking"
description: |-
  Look up networks (v3 networking API). Provide `name` for a single match, or omit to list all (optionally filtered by `domain`).
---

# mcs_network (Data Source)

Look up networks (v3 networking API). Provide `name` for a single match, or omit to list all (optionally filtered by `domain`).

## Example Usage

```hcl
data "mcs_network" "production" {
  name = "001.2002.USR"
}

output "network_vlan" {
  value = data.mcs_network.production.vlan_id
}
```

## Attributes

| Attribute     | Type   | Mode     | Description |
|--------------|--------|----------|-------------|
| `name`       | String | Optional | Exact network name to look up. |
| `domain`     | Number | Optional/Computed | Filter by domain ID; set to the matched network's domain. |
| `id`         | String | Computed | Network UUID. |
| `domain_name` | String | Computed | Name of the network's domain (from `domain_detail`). |
| `domain_vdom_id` | String | Computed | VDOM ID of the network's domain (from `domain_detail`). |
| `description` | String | Computed | Network description. |
| `ipv4_prefix` | String | Computed | IPv4 CIDR prefix. |
| `ipv4_address` | String | Computed | IPv4 address. |
| `vlan_id`    | Number | Computed | VLAN identifier (API field `vlanid`). |
| `customer`   | String | Computed | Customer that owns the network. |
| `customer_name` | String | Computed | Name of the customer that owns the network. |
| `type`       | String | Computed | Interface type: `vlan`, `tunnel`, `lacp` or `physical`. |
| `parent`     | String | Computed | Parent LACP interface. |
| `ipv6_prefix` | String | Computed | IPv6 prefix. |
| `ipv6_address` | String | Computed | IPv6 address. |
| `dhcp_server` | Bool  | Computed | Whether a DHCP server is enabled on the interface. |
| `dhcp_server_address_pool` | String | Computed | DHCP server address pool. |
| `dhcp_server_netmask` | String | Computed | DHCP server netmask. |
| `dhcp_server_dns_servers` | String | Computed | Comma-separated DNS servers handed out by the DHCP server. |
| `dhcp_server_lease_time` | Number | Computed | DHCP lease time. |
| `visible`    | Bool   | Computed | Whether the interface is visible to end users. |
| `deploy_flow` | String | Computed | Deploy flow that created the network: `v1` or `v2` (beta). |
| `datacenters` | List  | Computed | Datacenters the network was deployed to; each has `id` (Number) and `name` (String). |
| `leftover_steps` | List(String) | Computed | V2 only: steps whose configuration may still exist after a rollback or undeploy could not finish. |
| `fortimanager_install_pending_since` | String | Computed | V2 only: since when the network's FortiManager objects wait for an install of the domain's policy package; null once the FortiGate runs them. |
| `fortimanager_install_error` | String | Computed | V2 only: why the last check found the FortiGate not running the network's objects after the package was installed. |
| `created_at` | String | Computed | Creation timestamp. |
| `updated_at` | String | Computed | Last update timestamp. |
| `networks`   | List   | Computed | List of all networks (populated when `name` is not set). |

**Nested `networks` attributes:** the same attributes as the single-network attributes above (`id`, `name`, `domain`, `domain_name`, …, `updated_at`) — all Computed.
