---
page_title: "mcs_domain Data Source - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Look up firewall domains. Provide `name` for a single match, or omit to list all domains.
---

# mcs_domain (Data Source)

Look up firewall domains. Provide `name` for a single match, or omit to list all domains.

## Example Usage — Single lookup

```hcl
data "mcs_domain" "production" {
  name = "VDOM-PROD"
}

output "domain_uuid" {
  value = data.mcs_domain.production.uuid
}
```

## Example Usage — List all

```hcl
data "mcs_domain" "all" {}

output "all_domains" {
  value = data.mcs_domain.all.domains
}
```

## Attributes

| Attribute   | Type   | Mode     | Description |
|------------|--------|----------|-------------|
| `name`     | String | Optional | Exact domain name to look up. |
| `id`       | Number | Computed | Domain ID (set when a single domain is matched). |
| `uuid`     | String | Computed | Domain UUID. |
| `vdom_id`  | String | Computed | VDOM ID of the domain. |
| `description` | String | Computed | Domain description. |
| `adom`     | String | Computed | Administrative domain. |
| `zone_id`  | Number | Computed | ID of the associated network zone. |
| `zone_name` | String | Computed | Name of the associated network zone. |
| `customer` | String | Computed | Customer ID. |
| `customer_name` | String | Computed | Customer name. |
| `tag`      | List(String) | Computed | Tags of the domain, as returned by the API. |
| `comp_errors` | Number | Computed | Number of compliance errors. |
| `max_networks` | Number | Computed | Maximum number of networks in the domain. |
| `deploy_flow` | String | Computed | Deploy flow that deploys the domain: `v1` (deploy_vdom job) or `v2` (beta, step runner). |
| `state`    | String | Computed | Lifecycle state: `deploying`, `planned`, `active`, `rolling_back`, `rollback_failed`, `undeploying` or `undeploy_failed`. V1 domains are always `active`. |
| `domains`  | List   | Computed | List of all domains (populated when `name` is not set). |

**Nested `domains` attributes:** `id`, `uuid`, `vdom_id`, `name`, `description`, `adom`, `zone_id`, `zone_name`, `customer`, `customer_name`, `tag`, `comp_errors`, `max_networks`, `deploy_flow`, `state` — all Computed.

> The former `zone` attribute was removed: the API now returns the zone as an object, exposed as `zone_id` and `zone_name`.

> `tag` changed from `List(Number)` to `List(String)` because the API returns tags as strings.
