---
page_title: "mcs_csv_server Data Source - mcs"
subcategory: "Load Balancing"
description: |-
  Look up content switching virtual servers. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_csv_server (Data Source)

Look up content switching virtual servers. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_csv_server" "frontend" {
  name = "web-csvserver"
}
```

## Attributes

| Attribute      | Type         | Mode     | Description |
|---------------|-------------|----------|-------------|
| `name`        | String       | Optional | Exact CS vServer name. |
| `id`          | String       | Optional/Computed | CS vServer UUID. |
| `ufname`      | String       | Computed | User-friendly name. |
| `ipaddress`   | String       | Computed | UUID of the associated PublicIPAddress. |
| `port`        | Number       | Computed | Listening port. |
| `type`        | String       | Computed | Protocol type. |
| `policies`    | List(String) | Computed | CS policy IDs. |
| `certificate` | List(String) | Computed | SSL certificate IDs. |
| `ca_certificate` | List(String) | Computed | CA certificate IDs used to verify client certificates. |
| `clientauth`  | Bool         | Computed | Whether client certificate authentication is enabled. |
| `clientcert`  | String       | Computed | Client certificate requirement: `Mandatory` or `Optional`. |
| `customer`    | String       | Computed | Customer identifier. |
| `loadbalancer` | String      | Computed | UUID of the load balancer. |
| `csv_servers` | List         | Computed | All CS vServers (populated when neither `name` nor `id` is set). Items have the same attributes as above. |
