---
page_title: "mcs_integration Data Source - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Look up integrations. Provide `name` (exact match) or `id` for a single integration, or omit both to list all of them. Credentials are never returned.
---

# mcs_integration (Data Source)

Look up integrations. Provide `name` (exact match) or `id` for a single integration, or omit both to list all of them. Credentials are never returned.

## Example Usage

```hcl
data "mcs_integration" "prtg" {
  name = "prtg-prod"
}
```

## Attributes

| Attribute       | Type   | Mode              | Description |
|-----------------|--------|-------------------|-------------|
| `id`            | Number | Optional/Computed | Integration ID. |
| `name`          | String | Optional/Computed | Exact integration name. |
| `type_name`     | String | Computed          | Name of the integration type. |
| `subtype`       | String | Computed          | Integration subtype. |
| `baseurl`       | String | Computed          | Base URL of the external system. |
| `username`      | String | Computed          | Username used to authenticate. |
| `secret_expire` | String | Computed          | Expiry timestamp of the credentials. |
| `icon`          | String | Computed          | Icon of the integration. |
| `has_secret`    | Bool   | Computed          | Whether a secret is stored. |
| `has_key`       | Bool   | Computed          | Whether a key is stored. |
| `platform_shared` | Bool | Computed          | Whether the integration is shared platform-wide with other tenants. |
| `read_only`     | Bool   | Computed          | Whether the integration is read-only for the caller, e.g. a platform-shared integration of another tenant. |
| `config`        | Map of String | Computed   | Extra integration settings as key/value pairs. |
| `kubernetes_endpoint` | Object | Computed    | Kubernetes endpoint settings of a vSphere integration (`credentials_key`, `content_library`, `data_disk_max_size`, `nodestorage_step`); null when not set. |
| `integrations`  | List   | Computed          | All integrations (populated when neither `name` nor `id` is set). |
