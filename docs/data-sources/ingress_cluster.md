---
page_title: "mcs_ingress_cluster Data Source - mcs"
subcategory: "Secure Ingress"
description: |-
  Look up secure ingress clusters. Provide `name` or `id` for a single match; otherwise all clusters matching the optional `ipaddress`, `sla` and `bandwidth` filters are listed. `name` is sent as the `name__icontains` filter and then matched exactly.
---

# mcs_ingress_cluster (Data Source)

Look up secure ingress clusters. Provide `name` or `id` for a single match; otherwise all clusters matching the optional `ipaddress`, `sla` and `bandwidth` filters are listed. `name` is sent as the `name__icontains` filter and then matched exactly.

## Example Usage

```hcl
data "mcs_ingress_cluster" "webshop" {
  name = "webshop"
}

data "mcs_ingress_cluster" "gold" {
  sla = "gold"
}
```

## Attributes

| Attribute | Type | Mode | Description |
|-----------|------|------|-------------|
| `name` | String | Optional/Computed | Exact ingress cluster name. |
| `id` | String | Optional/Computed | Ingress cluster UUID. |
| `sla` | String | Optional/Computed | Filter by service level (`bronze`, `silver`, `gold`, `platinum`); the cluster's SLA for a single lookup. |
| `bandwidth` | Number | Optional/Computed | Filter by bandwidth in Mbps (`50`, `100`, `500`, `1000`, `5000`); the cluster's bandwidth for a single lookup. |
| `ipaddress` | String | Optional/Computed | Filter by public IP address UUID; the cluster's public IP UUID for a single lookup. |
| `customer` | String | Computed | Customer identifier. |
| `firewall` | String | Computed | UUID of the secure ingress firewall. |
| `slug` | String | Computed | URL-safe identifier derived by the API. |
| `state` | String | Computed | Synchronisation state: `synced`, `unsynced`, `error` or `deleted`. |
| `reverse_proxy` | Number | Computed | ID of the reverse proxy integration. |
| `reverse_proxy_name` | String | Computed | Name of the reverse proxy integration. |
| `ipaddress_address` | String | Computed | The public IP address the cluster listens on. |
| `ipaddress_type` | String | Computed | Type of the public IP address. |
| `created_at_timestamp` | String | Computed | Time when the cluster was created. |
| `updated_at_timestamp` | String | Computed | Time when the cluster was last updated. |
| `created_by_user` | Number | Computed | ID of the user who created the cluster. |
| `updated_by_user` | Number | Computed | ID of the user who last updated the cluster. |
| `ingress_clusters` | List | Computed | All matching ingress clusters (populated when neither `name` nor `id` is set). |

**Nested `ingress_clusters` attributes:** the same attributes as the single-lookup attributes above (`id`, `name`, `sla`, `bandwidth`, `customer`, `ipaddress`, `firewall`, `slug`, `state`, `reverse_proxy`, `reverse_proxy_name`, `ipaddress_address`, `ipaddress_type`, `created_at_timestamp`, `updated_at_timestamp`, `created_by_user`, `updated_by_user`) — all Computed.
