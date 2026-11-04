---
page_title: "mcs_network_dhcp_lease Resource - mcs"
subcategory: "Networking"
description: |-
  Manages a static DHCP lease (FortiOS reserved address) on a network with a DHCP server.
---

# mcs_network_dhcp_lease (Resource)

Manages a static DHCP lease: a FortiOS reserved address that always gives one MAC address the same IP. The network must have a DHCP server (see `mcs_network_dhcp_server`).

Every create, update and delete queues one DHCP change job and waits until it has finished. If the job fails, MCS rolls the change back on the firewall, and the error shows the failed step and the rollback error. The create request is sent exactly once (it is never retried automatically, so a flaky connection cannot queue a second lease).

**Ordering.** Set `network` to `mcs_network_dhcp_server.<name>.network` rather than to the network's ID. Terraform then creates the lease after the server and destroys it before the server, which MCS requires: it refuses to disable a server that still has leases.

**MAC notation.** `mac` accepts any common notation (`00:50:56:aa:bb:01`, `00-50-56-AA-BB-01`, `0050.56aa.bb01`). MCS stores it lower case and colon separated. State keeps the notation you configured, so it never shows a permanent diff, and rewriting the same address in another notation updates state without queueing a change.

**Validation.** MCS checks a lease before queueing it: the IP must be inside the network's prefix and must not be its network, broadcast or gateway address. The IP and MAC must also be unique on the network, counting leases that are still queued. A rejected value is reported per field (HTTP 400).

## Example Usage

```hcl
data "mcs_network" "office" {
  name = "NET-OFFICE"
}

resource "mcs_network_dhcp_server" "office" {
  network     = data.mcs_network.office.id
  pool_start  = "10.20.30.100"
  pool_end    = "10.20.30.200"
  dns_servers = ["10.1.1.1"]
  lease_time  = 86400
}

locals {
  reservations = {
    printer = { ip = "10.20.30.11", mac = "00:50:56:aa:bb:01" }
    scanner = { ip = "10.20.30.12", mac = "00-50-56-AA-BB-02" }
  }
}

resource "mcs_network_dhcp_lease" "office" {
  for_each = local.reservations

  # Referring to the server orders the lease after its creation and before its removal.
  network     = mcs_network_dhcp_server.office.network
  ip          = each.value.ip
  mac         = each.value.mac
  description = each.key
}
```

## Attributes

| Attribute     | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `network`     | String | **Yes**  | UUID of the network; it must have a DHCP server. Changing it recreates the lease (a lease never moves to another network). |
| `ip`          | String | **Yes**  | IPv4 address to reserve. |
| `mac`         | String | **Yes**  | MAC address of the client, in any common notation. |
| `description` | String | No       | What the reservation is for (max 255 characters). Defaults to `""`. |
| `timeouts`    | Block  | No       | `create`, `update` and `delete` durations (e.g. `"30m"`). All default to 20 minutes. |

`ip`, `mac` and `description` are changed in place.

**Read-only attributes:**

| Attribute    | Type   | Description |
|--------------|--------|-------------|
| `id`         | String | UUID of the lease. |
| `created_at` | String | Creation timestamp. |
| `updated_at` | String | Last update timestamp. |

## Import

Leases the device sync found on the firewall appear in MCS with their own UUID; find them with the `mcs_network_dhcp_leases` data source.

```shell
terraform import mcs_network_dhcp_lease.printer <lease-uuid>
```

After an import `mac` is in MCS's notation. If your configuration writes it differently, the next apply updates state without queueing a change.
