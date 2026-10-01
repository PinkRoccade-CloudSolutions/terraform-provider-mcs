<!-- TOC entry (Resources > Networking): -->
<!--     - [mcs_ippool](#mcs_ippool-resource) -->

#### mcs_ippool (Resource)

Manages an IP pool from which public IP addresses are allocated.

##### Example

```hcl
resource "mcs_ippool" "nat" {
  name     = "NAT pool"
  subnet   = "203.0.113.0/28"
  type     = "nat"
  customer = mcs_customer.example.id
}

output "nat_pool_free_ips" {
  value = mcs_ippool.nat.free_ips
}
```

##### Attributes

| Attribute  | Type   | Required | Description |
|-----------|--------|----------|-------------|
| `name`    | String | **Yes**  | Pool name (1–100 characters). |
| `subnet`  | String | **Yes**  | Pool subnet in CIDR notation (max 50). Changing it forces a new resource, because the addresses allocated from the pool depend on it. |
| `type`    | String | No       | Pool type: `nat`, `vip`, `loadbalancer`, or `""`. Computed by the API when omitted. |
| `customer` | String | No      | Customer owning the pool. Computed by the API when omitted. |

**Read-only attributes:**

| Attribute   | Type   | Description |
|------------|--------|-------------|
| `id`       | String | UUID of the IP pool. |
| `total_ips` | String | Total number of addresses in the pool. |
| `free_ips` | String | Number of free addresses in the pool. |

##### Import

```shell
terraform import mcs_ippool.nat <pool-uuid>
```
