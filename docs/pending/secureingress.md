<!--
Pending docs for mcs_secureingress_firewall and mcs_ingress_cluster. Merge into docs/index.md:
- TOC, under "Data Sources" (after mcs_site_to_site_vpn):
    - [mcs_secureingress_firewall](#mcs_secureingress_firewall-data-source)
    - [mcs_ingress_cluster](#mcs_ingress_cluster-data-source)
- TOC, under "Resources" (new group after "Networking"):
    - [Secure Ingress](#secure-ingress)
      - [mcs_secureingress_firewall](#mcs_secureingress_firewall)
      - [mcs_ingress_cluster](#mcs_ingress_cluster)
- Data source sections: at the end of the Data Sources part.
- Resource sections: a new "### Secure Ingress" group after "### Networking".
-->

### mcs_secureingress_firewall (Data Source)

Look up secure ingress (XDP) firewalls. Provide `name` or `id` for a single match, or omit both to list all. `name` is sent as the `name__icontains` filter and then matched exactly.

#### Example

```hcl
data "mcs_secureingress_firewall" "edge" {
  name = "edge-filter"
}
```

#### Attributes

| Attribute | Type | Mode | Description |
|-----------|------|------|-------------|
| `name` | String | Optional/Computed | Exact firewall name. |
| `id` | String | Optional/Computed | Firewall UUID. |
| `description` | String | Computed | Description of the firewall. |
| `filter_json` | String (JSON) | Computed | Firewall filter definition as a JSON document. Use `jsondecode()` to read it. |
| `customer` | String | Computed | Customer identifier. |
| `created_at_timestamp` | String | Computed | Time when the firewall was created. |
| `updated_at_timestamp` | String | Computed | Time when the firewall was last updated. |
| `created_by_user` | Number | Computed | ID of the user who created the firewall. |
| `updated_by_user` | Number | Computed | ID of the user who last updated the firewall. |
| `secureingress_firewalls` | List | Computed | All secure ingress firewalls (populated when neither `name` nor `id` is set). |

**Nested `secureingress_firewalls` attributes:** `id`, `name`, `description`, `filter_json`, `customer`, `created_at_timestamp`, `updated_at_timestamp` (String), `created_by_user`, `updated_by_user` (Number) — all Computed.

---

### mcs_ingress_cluster (Data Source)

Look up secure ingress clusters. Provide `name` or `id` for a single match; otherwise all clusters matching the optional `ipaddress`, `sla` and `bandwidth` filters are listed. `name` is sent as the `name__icontains` filter and then matched exactly.

#### Example

```hcl
data "mcs_ingress_cluster" "webshop" {
  name = "webshop"
}

data "mcs_ingress_cluster" "gold" {
  sla = "gold"
}
```

#### Attributes

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

---

### Secure Ingress

#### mcs_secureingress_firewall

Manages a secure ingress (XDP) firewall filter. Referenced by `mcs_ingress_cluster.firewall`.

The API stores `filter_json` as a JSON-encoded string. Write it with `jsonencode()`; differences in whitespace or key order between your configuration and the API's response do not cause a diff.

##### Example

```hcl
resource "mcs_secureingress_firewall" "edge" {
  name        = "edge-filter"
  description = "Blocks unwanted traffic"
  customer    = mcs_customer.example.id
  filter_json = jsonencode({
    allow = ["10.0.0.0/8"]
  })
}
```

##### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | **Yes** | Name of the firewall (max 200 characters). |
| `filter_json` | String (JSON) | **Yes** | Firewall filter definition as a JSON document. |
| `customer` | String | **Yes** | Customer identifier. Changing this forces a new resource. |
| `description` | String | No | Description of the firewall (max 200 characters). |

**Read-only attributes:** `id`, `created_at_timestamp`, `updated_at_timestamp` (String), `created_by_user`, `updated_by_user` (Number).

##### Import

```shell
terraform import mcs_secureingress_firewall.edge <firewall-uuid>
```

---

#### mcs_ingress_cluster

Manages a secure ingress cluster: a public IP address protected by a secure ingress (XDP) firewall and fronted by a reverse proxy. Instances, routes and upstream targets are not managed by this resource.

##### Example

```hcl
resource "mcs_public_ip_address" "ingress" {
  type     = "secureingress"
  customer = mcs_customer.example.id
}

resource "mcs_ingress_cluster" "webshop" {
  name      = "webshop"
  sla       = "gold"
  bandwidth = 1000
  customer  = mcs_customer.example.id
  ipaddress = mcs_public_ip_address.ingress.id
  firewall  = mcs_secureingress_firewall.edge.id
}
```

##### Attributes

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | String | **Yes** | Name of the ingress cluster (max 200 characters). |
| `customer` | String | **Yes** | Customer identifier. Changing this forces a new resource. |
| `ipaddress` | String | **Yes** | UUID of the public IP address (`mcs_public_ip_address`) the cluster listens on. |
| `firewall` | String | **Yes** | UUID of the secure ingress firewall (`mcs_secureingress_firewall`). |
| `sla` | String | No | Service level: `bronze`, `silver`, `gold` or `platinum`. Server default when unset. |
| `bandwidth` | Number | No | Bandwidth in Mbps: `50`, `100`, `500`, `1000` or `5000`. Server default when unset. |

**Read-only attributes:** `id`, `slug`, `state`, `reverse_proxy_name`, `ipaddress_address`, `ipaddress_type`, `created_at_timestamp`, `updated_at_timestamp` (String), `reverse_proxy`, `created_by_user`, `updated_by_user` (Number).

##### Import

```shell
terraform import mcs_ingress_cluster.webshop <ingress-cluster-uuid>
```

---
