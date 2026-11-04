---
page_title: "mcs_public_ip_address Resource - mcs"
subcategory: "Networking"
description: |-
  Manages a public IP address allocation from an IP pool.
---

# mcs_public_ip_address (Resource)

Manages a public IP address allocation from an IP pool.

## Example Usage

```hcl
data "mcs_ippool" "nat_pool" {
  name = "NAT Pool"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_public_ip_address" "web" {
  pool        = data.mcs_ippool.nat_pool.id
  description = "Web server public IP"
  type        = "nat"
  customer    = data.mcs_customer.example.id
}
```

## Attributes

| Attribute    | Type   | Required | Description |
|-------------|--------|----------|-------------|
| `pool`      | String | No       | UUID of the IP pool to allocate from. Computed by the API when omitted. Changing it forces a new resource. |
| `ip_address` | String | No      | Specific public IP address to claim. Assigned by the API when omitted. |
| `description` | String | No     | Description of the public IP address (max 255). Defaults to `""`; removing it clears the description. |
| `type`      | String | No       | Type: `nat`, `vip`, `loadbalancer`, or `secureingress`. |
| `customer`  | String | No       | Customer identifier. Computed by the API when omitted. |

**Read-only attributes:**

| Attribute    | Type   | Description |
|-------------|--------|-------------|
| `id`        | String | UUID of the public IP address. |
| `status`    | String | Allocation status: `available`, `assigned`, or `reserved`. |

## Import

```shell
terraform import mcs_public_ip_address.web <public-ip-uuid>
```
