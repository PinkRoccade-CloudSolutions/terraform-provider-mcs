---
page_title: "mcs_customer Data Source - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Look up customers. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_customer (Data Source)

Look up customers. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_customer" "prod" {
  name = "Production Customer"
}
```

## Attributes

| Attribute        | Type         | Mode     | Description |
|-----------------|-------------|----------|-------------|
| `name`          | String       | Optional | Exact customer name (sent as `name__icontains`, matched exactly client-side). |
| `id`            | String       | Optional/Computed | Customer ID. |
| `slug`          | String       | Computed | Short DNS label of the customer; prefixes its Kubernetes cluster names. |
| `contractid`    | String       | Computed | Contract identifier. |
| `sdm`           | Number       | Computed | Service Delivery Manager user ID; null when not set. |
| `admin_contacts` | List(Number) | Computed | Administrative contact IDs. |
| `tech_contacts` | List(Number) | Computed | Technical contact IDs. |
| `customers`     | List         | Computed | All customers (populated when neither `name` nor `id` is set). |
