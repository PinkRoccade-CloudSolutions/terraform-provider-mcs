---
page_title: "mcs_network_placements Data Source - mcs"
subcategory: "Networking"
description: |-
  Shows which datacenters a V2 deploy of the `mcs_network` resource would go to in a domain, for each `placement`. An empty `standard` list means the domain's zone is not ready for V2 deploys.
---

# mcs_network_placements (Data Source)

Shows which datacenters a V2 deploy of the `mcs_network` resource would go to in a domain, for each `placement`. An empty `standard` list means the domain's zone is not ready for V2 deploys.

## Example Usage

```hcl
data "mcs_domain" "production" {
  name = "production"
}

data "mcs_network_placements" "production" {
  domain_uuid = data.mcs_domain.production.uuid
}

output "kubernetes_datacenters" {
  value = data.mcs_network_placements.production.kubernetes[*].name
}
```

## Attributes

| Attribute                    | Type   | Mode         | Description |
|------------------------------|--------|--------------|-------------|
| `domain_uuid`                | String | **Required** | UUID of the domain (VDOM), e.g. `data.mcs_domain.x.uuid`. |
| `standard`                   | List   | Computed     | Datacenters for `placement = "standard"`; each has `id` (Number) and `name` (String). |
| `kubernetes`                 | List   | Computed     | Datacenters for `placement = "kubernetes"`: the standard ones plus the zone's additional datacenters. Same nested attributes. |
| `kubernetes_equals_standard` | Bool   | Computed     | Whether both placements deploy to the same datacenters. |
