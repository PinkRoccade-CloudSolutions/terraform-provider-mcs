---
page_title: "mcs_secureingress_route Data Source - mcs"
subcategory: "Secure Ingress"
description: |-
  List secure ingress routes, optionally filtered by `id`, `cluster` and exact `server_name`.
---

# mcs_secureingress_route (Data Source)

List secure ingress routes, optionally filtered by `id`, `cluster` and exact `server_name`.

## Example Usage

```hcl
data "mcs_ingress_cluster" "edge" {
  name = "edge"
}

data "mcs_secureingress_route" "shop" {
  cluster     = data.mcs_ingress_cluster.edge.id
  server_name = "shop.example.com"
}
```

## Attributes

| Attribute     | Type   | Mode     | Description |
|---------------|--------|----------|-------------|
| `id`          | String | Optional | Return only the route with this ID. |
| `cluster`     | String | Optional | Return only routes of this ingress cluster. |
| `server_name` | String | Optional | Return only the route with this exact server name. |
| `routes`      | List   | Computed | Matching routes, each with `id`, `cluster`, `server_name`, `target_kind`, `virtualmachine`, `interface`, `kubernetes_cluster`, `upstream_address`, `port`, `scheme`, `path`, `ssl`, `upstream`, `certificate_common_name`, `certificate_issuer` and `certificate_valid_to_timestamp`. |
