---
page_title: "mcs_network_dhcp_server Resource - mcs"
subcategory: "Networking"
description: |-
  Manages the DHCP server of a network (one per network) through the v3 networking API.
---

# mcs_network_dhcp_server (Resource)

Manages the DHCP server of a network through the v3 networking API. A network has at most one DHCP server, so the resource is keyed by the network's UUID.

Every create, update and delete queues one DHCP change job that configures the FortiGate and has FortiManager retrieve it; Terraform waits until the job has finished. If the job fails, MCS rolls the change back on the firewall, and the error shows the failed step and the rollback error. Destroying the resource disables the DHCP server.

**Static leases.** MCS refuses (HTTP 409) to disable a DHCP server while its network still has static leases. Set each `mcs_network_dhcp_lease`'s `network` to `mcs_network_dhcp_server.<name>.network`, as in the example below, so Terraform creates leases after the server and destroys them before it.

**Existing servers.** Creating the resource on a network that already runs a DHCP server fails, so Terraform never silently takes over a server. Import it instead.

**Concurrency.** MCS queues changes to the same network behind each other, so many leases and the server can be applied in parallel. A network that is not active, or has a deploy or teardown queued, is refused with HTTP 409; the error is reported as is and not retried.

## Example Usage

```hcl
data "mcs_network" "office" {
  name = "NET-OFFICE"
}

resource "mcs_network_dhcp_server" "office" {
  network     = data.mcs_network.office.id
  pool_start  = "10.20.30.100"
  pool_end    = "10.20.30.200"
  dns_servers = ["10.1.1.1", "10.1.1.2"]
  lease_time  = 86400
}

# Refer to the server's `network`, not the network itself, so Terraform creates
# leases after the server and destroys them before it (MCS refuses to disable a
# DHCP server that still has static leases).
resource "mcs_network_dhcp_lease" "printer" {
  network     = mcs_network_dhcp_server.office.network
  ip          = "10.20.30.11"
  mac         = "00:50:56:aa:bb:01"
  description = "Printer 2nd floor"
}
```

## Attributes

| Attribute     | Type         | Required | Description |
|---------------|--------------|----------|-------------|
| `network`     | String       | **Yes**  | UUID of the network to run the DHCP server on. Changing it disables the server on the old network and enables it on the new one. |
| `pool_start`  | String       | **Yes**  | First address of the dynamic pool. Must be inside the network's prefix and not its network, broadcast or gateway address. |
| `pool_end`    | String       | **Yes**  | Last address of the dynamic pool. Must not be before `pool_start`, and the pool must not contain the gateway. |
| `dns_servers` | List(String) | **Yes**  | DNS servers handed out to clients: one to four IPv4 addresses. |
| `lease_time`  | Number       | **Yes**  | Lease time in seconds (300–8604800). |
| `timeouts`    | Block        | No       | `create`, `update` and `delete` durations (e.g. `"30m"`). All default to 20 minutes. |

MCS validates the pool against the network's prefix before queueing anything; a rejected value is reported per field (HTTP 400).

**Read-only attributes:**

| Attribute         | Type   | Description |
|-------------------|--------|-------------|
| `id`              | String | UUID of the network (the DHCP server has no ID of its own). |
| `netmask`         | String | Netmask handed out to clients; follows the network's prefix. |
| `default_gateway` | String | Default gateway handed out to clients; the network's gateway. |

If the DHCP server is disabled outside Terraform, the next refresh removes it from state and the next apply enables it again.

## Import

```shell
terraform import mcs_network_dhcp_server.office <network-uuid>
```
