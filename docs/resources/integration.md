---
page_title: "mcs_integration Resource - mcs"
subcategory: "Tenant & Customer Management"
description: |-
  Manages an integration with an external system (e.g. PRTG) within the tenant. Set the type with either `type_id` or `catalog_key`. When you use `catalog_key`, the API resolves the type, subtype and icon from the integration catalog at creation.
---

# mcs_integration (Resource)

Manages an integration with an external system (e.g. PRTG) within the tenant. Set the type with either `type_id` or `catalog_key`. When you use `catalog_key`, the API resolves the type, subtype and icon from the integration catalog at creation.

`secret` and `key` are write-only and sensitive. The API never returns them, so Terraform keeps the configured values in state and cannot detect credential changes made outside Terraform. Use `has_secret`/`has_key` to see whether a credential is stored.

## Example Usage

```hcl
variable "prtg_password" {
  type      = string
  sensitive = true
}

resource "mcs_integration" "prtg" {
  name        = "prtg-prod"
  catalog_key = "prtg"
  baseurl     = "https://prtg.example.com"
  username    = "svc-mcs"
  secret      = var.prtg_password
}
```

## Attributes

| Attribute       | Type   | Required | Description |
|-----------------|--------|----------|-------------|
| `name`          | String | **Yes**  | Integration name. Cannot contain special characters or a dot. |
| `baseurl`       | String | **Yes**  | Base URL of the external system. |
| `type_id`       | Number | No       | ID of the integration type. Write-only. |
| `catalog_key`   | String | No       | Integration catalog entry. Write-only; changing it forces replacement. |
| `subtype`       | String | No       | Integration subtype. Resolved from `catalog_key` when omitted. |
| `username`      | String | No       | Username used to authenticate. |
| `secret`        | String | No       | Secret or password. Sensitive, write-only. |
| `key`           | String | No       | API key. Sensitive, write-only. |
| `secret_expire` | String | No       | Expiry timestamp of the credentials (RFC 3339). |
| `kubernetes_endpoint` | Object | No | Kubernetes endpoint settings of a vSphere integration; see below. Removing it clears the settings. |

**`kubernetes_endpoint` attributes:**

| Attribute            | Type   | Required | Description |
|---------------------|--------|----------|-------------|
| `credentials_key`    | String | **Yes**  | Key of the vCenter credentials in the Azure DevOps Library. Not a secret. |
| `content_library`    | String | **Yes**  | vSphere Content Library with the node images. |
| `data_disk_max_size` | Number | No       | Largest node storage disk (GB) on this endpoint; at least 500. Computed from the API when omitted. |
| `nodestorage_step`   | Number | No       | Node storage sizes (GB) are a multiple of this. Computed from the API when omitted. |

```hcl
variable "vcenter_password" {
  type      = string
  sensitive = true
}

resource "mcs_integration" "vcenter" {
  name        = "vcenter01"
  catalog_key = "vsphere"
  baseurl     = "https://vcenter01.example.com"
  username    = "svc-mcs"
  secret      = var.vcenter_password

  kubernetes_endpoint = {
    credentials_key = "vcenter01-k8s"
    content_library = "k8s-node-images"
  }
}
```

**Read-only attributes:** `id` (Number), `type_name` (String), `icon` (String), `has_secret` (Bool), `has_key` (Bool).


## Import

The import ID is the numeric ID of the integration. `secret` and `key` are never returned by the API, so they stay empty after import until they are set in the configuration.

```shell
terraform import mcs_integration.prtg <integration-id>
```
