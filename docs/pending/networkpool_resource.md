<!-- TOC entry (Resources > Networking): -->
<!--     - [mcs_networkpool](#mcs_networkpool-resource) -->

#### mcs_networkpool (Resource)

Manages a network pool that end users can pick networks from.

##### Example

```hcl
resource "mcs_networkpool" "lan" {
  name        = "Production LAN Pool"
  network     = "10.0.0.0/8"
  description = "Internal LAN networks"
  type        = "lan"
  enabled     = true
}
```

##### Attributes

All arguments are optional. Omitted values are taken from the API response; set a string to `""` to clear it.

| Attribute     | Type   | Required | Description |
|--------------|--------|----------|-------------|
| `name`       | String | No       | Pool name (max 255). |
| `network`    | String | No       | Pool network in CIDR notation, e.g. `10.0.0.0/8` (max 255). |
| `description` | String | No      | Description to help end users pick a pool (max 2048). |
| `type`       | String | No       | Pool type: `lan`, `wan`, `transit`, or `""`. |
| `enabled`    | Bool   | No       | Whether the pool is usable by end users. |

**Read-only attributes:** `id` (String) — UUID of the network pool.

##### Import

```shell
terraform import mcs_networkpool.lan <pool-uuid>
```
