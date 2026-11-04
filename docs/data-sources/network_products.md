---
page_title: "mcs_network_products Data Source - mcs"
subcategory: "Networking"
description: |-
  Lists the products a network can be created under — the allowed values of the `product` argument of the `mcs_network` resource. Each product owns a PRODID band.
---

# mcs_network_products (Data Source)

Lists the products a network can be created under — the allowed values of the `product` argument of the `mcs_network` resource. Each product owns a PRODID band.

## Example Usage

```hcl
data "mcs_network_products" "all" {}

output "network_products" {
  value = data.mcs_network_products.all.names
}
```

## Attributes

| Attribute | Type         | Mode     | Description |
|-----------|--------------|----------|-------------|
| `names`   | List(String) | Computed | Names of the available network products. |
