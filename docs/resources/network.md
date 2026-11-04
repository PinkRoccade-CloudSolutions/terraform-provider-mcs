---
page_title: "mcs_network Resource - mcs"
subcategory: "Networking"
description: |-
  Deploys a network (VLAN, prefix and interface) in a domain through the v3 networking API.
---

# mcs_network (Resource)

Deploys a network (VLAN, prefix and interface) in a domain through the v3 networking API.

Creating the resource queues a deploy job and waits until it has finished; destroying it queues a teardown and waits until the network is gone. The deploy request is sent exactly once (it is never retried automatically, so a flaky connection cannot queue a second network). If the job fails after MCS has already written the network, the network is kept in state as tainted and is replaced on the next apply.

Networks cannot be changed in place: changing any argument destroys the network and deploys a new one. The one exception is `dry_run`: changing it from `true` to `false` deploys the planned network in place.

**Deploy flows.** `deploy_flow = "v1"` (the default) is the classic deploy. `deploy_flow = "v2"` (beta) deploys step by step to the datacenters chosen by `placement`, and enables `dry_run`. If a V2 deploy fails, MCS rolls back the steps it already ran; the error then reports the rollback state, the failed step and any steps whose configuration may remain (also exposed afterwards as `leftover_steps`). Caveats of a deploy or undeploy that otherwise succeeded, such as a deferred FortiManager install, are shown as Terraform warnings.

**Kubernetes placement.** With V2, `placement = "standard"` (the default) deploys to the zone's default datacenters and `placement = "kubernetes"` also deploys to the zone's additional datacenters. Use the `mcs_network_placements` data source to see which datacenters each placement resolves to before deploying.

**Dry run.** With V2, `dry_run = true` checks every system without changing anything: MCS reserves the VLAN and prefix and keeps the network *planned*. The changes the deploy would make are listed in `dry_run_changes`; changes with action `conflict` are also reported as a warning, since the deploy would fail on them. Change `dry_run` to `false` to deploy the planned network in place (the `update` timeout applies). If that deploy fails it is rolled back, the network stays planned and `dry_run` stays `true` in state, so the next apply retries the deploy. Changing `dry_run` from `false` to `true` destroys the deployed network and creates a new planned one.

## Example Usage

```hcl
data "mcs_domain" "production" {
  name = "production"
}

data "mcs_networkpool" "lan" {
  name = "Production LAN Pool"
}

resource "mcs_network" "web" {
  domain_uuid     = data.mcs_domain.production.uuid
  product         = "DMZ"
  bitmask         = 26
  description     = "Web tier"
  network_pool_id = data.mcs_networkpool.lan.id

  timeouts {
    create = "45m"
  }
}

output "web_vlan" {
  value = mcs_network.web.vlan_id
}
```

## Example Usage — V2 dry run on Kubernetes placement, then deploy

```hcl
data "mcs_domain" "production" {
  name = "production"
}

data "mcs_networkpool" "lan" {
  name = "LAN Pool"
}

data "mcs_network_placements" "production" {
  domain_uuid = data.mcs_domain.production.uuid
}

resource "mcs_network" "k8s" {
  domain_uuid     = data.mcs_domain.production.uuid
  product         = "DMZ"
  bitmask         = 26
  description     = "Kubernetes ingress"
  network_pool_id = data.mcs_networkpool.lan.id
  deploy_flow     = "v2"
  placement       = "kubernetes"

  # Apply with true to see what the deploy would do, then set to false to deploy in place.
  dry_run = true
}

output "k8s_datacenters" {
  value = data.mcs_network_placements.production.kubernetes[*].name
}

output "k8s_planned_changes" {
  value = mcs_network.k8s.dry_run_changes
}
```

## Attributes

| Attribute         | Type   | Required | Description |
|-------------------|--------|----------|-------------|
| `domain_uuid`     | String | **Yes**  | UUID of the domain (VDOM) to attach the network to, e.g. `data.mcs_domain.x.uuid`. |
| `product`         | String | **Yes**  | Product that owns the network; picks the PRODID band. One of `Algemeen`, `CareCTRL`, `Caress`, `CreAim`, `DMZ`, `Geniq - Planywhere`, `MyHealthOnline`, `NPQ`, `PQConnect`, `PRLG`, `Quarant`, `ValueCare`, `Windex` (see `mcs_network_products`). |
| `bitmask`         | Number | **Yes**  | Prefix length of the network (24–29). |
| `description`     | String | **Yes**  | What the network is for (max 4096 characters). |
| `network_pool_id` | String | No       | UUID of the root network pool to allocate the prefix from. Conflicts with `prefix`. |
| `prefix`          | String | No       | Explicit network address to use instead of allocating from a pool, e.g. `10.0.0.0`. Conflicts with `network_pool_id`. |
| `deploy_flow`     | String | No       | `v1` (default) or `v2` (beta). Undeploy follows the same flow. |
| `placement`       | String | No       | V2 only: `standard` (default for V2) or `kubernetes`. Null with V1. |
| `dry_run`         | Bool   | No       | V2 only: plan the deploy without changing anything (default `false`). See *Dry run* above. |
| `timeouts`        | Block  | No       | `create`, `update` (deploying a planned network) and `delete` durations (e.g. `"45m"`). All default to 30 minutes. |

To manage the network's DHCP server and static leases, use `mcs_network_dhcp_server` and `mcs_network_dhcp_lease`; the `dhcp_server_*` attributes below are read-only.

All arguments except `timeouts` force a new network when changed, apart from `dry_run` going from `true` to `false`, which deploys in place. Setting `placement` or `dry_run = true` without `deploy_flow = "v2"` is rejected at plan time.

**Read-only attributes:**

| Attribute         | Type   | Description |
|-------------------|--------|-------------|
| `id`              | String | UUID of the network. |
| `name`            | String | Network name, assigned by MCS. |
| `domain_id`       | Number | Integer ID of the network's domain. |
| `domain_name`     | String | Name of the network's domain. |
| `domain_vdom_id`  | String | VDOM ID of the network's domain. |
| `customer`        | String | Customer that owns the network. |
| `customer_name`   | String | Name of the customer that owns the network. |
| `type`            | String | Interface type: `vlan`, `tunnel`, `lacp` or `physical`. |
| `parent`          | String | Parent LACP interface. |
| `vlan_id`         | Number | VLAN ID assigned to the network. |
| `ipv4_prefix`     | String | IPv4 prefix (CIDR). |
| `ipv4_address`    | String | IPv4 address. |
| `ipv6_prefix`     | String | IPv6 prefix. |
| `ipv6_address`    | String | IPv6 address. |
| `dhcp_server`     | Bool   | Whether a DHCP server is enabled on the interface. |
| `dhcp_server_address_pool` | String | DHCP server address pool. |
| `dhcp_server_netmask` | String | DHCP server netmask. |
| `dhcp_server_dns_servers` | String | Comma-separated DNS servers handed out by the DHCP server. |
| `dhcp_server_lease_time` | Number | DHCP lease time. |
| `visible`         | Bool   | Whether the interface is visible to end users. |
| `datacenters`     | List   | Datacenters the network was deployed to; each has `id` (Number) and `name` (String). |
| `leftover_steps`  | List(String) | V2 only: steps whose configuration may still exist after a rollback or undeploy could not finish, in the order they still need reversing. |
| `fortimanager_install_pending_since` | String | V2 only: since when the network's FortiManager objects wait for an install of the domain's policy package; null once the FortiGate runs them. |
| `fortimanager_install_error` | String | V2 only: why the last check found the FortiGate not running the network's objects after the package was installed. |
| `dry_run_changes` | List   | Changes the dry run reported, captured when the planned network was created; empty once deployed. Each has `step`, `system`, `datacenter` (null when not datacenter-specific), `action` (`create`, `update`, `present` or `conflict`), `object` and `detail`. |
| `created_at`      | String | Creation timestamp. |
| `updated_at`      | String | Last update timestamp. |

## Import

```shell
terraform import mcs_network.web <network-uuid>
```

The API does not return `product`, `network_pool_id`, `prefix` or `placement`, so after an import they are taken from your configuration on the next apply as an in-place update, without recreating the network. `domain_uuid` is looked up from the network's domain, `bitmask` is derived from `ipv4_prefix`, `deploy_flow` is read from the network, and `dry_run` is `true` when the network is still planned. `dry_run_changes` is empty after an import.
