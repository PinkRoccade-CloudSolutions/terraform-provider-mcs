---
page_title: "mcs_secureingress_route Resource - mcs"
subcategory: "Secure Ingress"
description: |-
  Manages a route on a secure ingress cluster. A route publishes a hostname and proxies it to an upstream: a virtual machine interface (`vm`), a Kubernetes cluster's load balancer (`kubernetes`), or a fixed IP address (`address`). The target kind is inferred from the target fields you set. A route can carry BunkerWeb setting overrides, extra locations (paths with their own upstream), WAF exceptions and a custom certificate. Without a custom certificate it uses Let's Encrypt.
---

# mcs_secureingress_route (Resource)

Manages a route on a secure ingress cluster. A route publishes a hostname and proxies it to an upstream: a virtual machine interface (`vm`), a Kubernetes cluster's load balancer (`kubernetes`), or a fixed IP address (`address`). The target kind is inferred from the target fields you set. A route can carry BunkerWeb setting overrides, extra locations (paths with their own upstream), WAF exceptions and a custom certificate. Without a custom certificate it uses Let's Encrypt.

Writes fail until the ingress cluster has been deployed. In that case the provider reports **"Ingress cluster not deployed yet"**; wait for the cluster to sync and apply again.

> The route API is not yet part of the published MCS API specification; this resource follows the MCS backend implementation.

Behaviour to be aware of:

- **Settings.** All values are strings, and booleans are written `"yes"`/`"no"`. Keys you remove from `settings` are reset to their default. The API drops values that equal the default and normalises some values, such as list order, case and duplicates. Terraform keeps your configured value in those cases, so no diff appears. As a trade-off, resetting a configured setting to its default outside Terraform is not detected.
- **Secret settings.** Write-only settings such as `AUTH_BASIC_PASSWORD` go in the sensitive `secret_settings` map. The API never returns them; `secrets_set` lists which ones are stored.
- **Locations** are matched to existing locations by `path`, so reordering them updates locations in place instead of recreating them. Locations that are left out are deleted.
- **WAF exceptions** are applied in list order and matched to existing exceptions by position. Raw rules (`kind = "raw"`) need the raw WAF rule permission.
- **Custom certificate.** The certificate is uploaded after the route is saved and re-uploaded when any input changes. Removing the block reverts to Let's Encrypt. The inputs are sensitive and kept in state because the API only returns certificate metadata. For binary PFX/P7B bundles, pass base64 content, e.g. `filebase64("bundle.pfx")`.

## Example Usage — VM upstream with settings, a location and a WAF exception

```hcl
variable "shop_basic_auth_password" {
  type      = string
  sensitive = true
}

data "mcs_ingress_cluster" "edge" {
  name = "edge"
}

data "mcs_virtualmachine" "web" {
  name = "web-server-01"
}

data "mcs_interface" "web" {
  name = "eth0"
}

resource "mcs_secureingress_route" "shop" {
  cluster        = data.mcs_ingress_cluster.edge.id
  server_name    = "shop.example.com"
  virtualmachine = data.mcs_virtualmachine.web.id
  interface      = data.mcs_interface.web.id
  port           = 8080

  settings = {
    USE_MODSECURITY = "yes"
    ALLOWED_METHODS = "GET|POST"
    USE_AUTH_BASIC  = "yes"
  }
  secret_settings = {
    AUTH_BASIC_PASSWORD = var.shop_basic_auth_password
  }

  locations = [{
    path             = "/api"
    upstream_address = "10.0.0.9"
    port             = 9000
  }]

  waf_exceptions = [{
    rules  = "942100"
    path   = "/search"
    target = "ARGS:q"
    reason = "Search box triggers SQLi false positives"
  }]
}
```

## Example Usage — Fixed address with a custom certificate

```hcl
data "mcs_ingress_cluster" "edge" {
  name = "edge"
}

resource "mcs_secureingress_route" "legacy" {
  cluster          = data.mcs_ingress_cluster.edge.id
  server_name      = "legacy.example.com"
  target_kind      = "address"
  upstream_address = "192.0.2.10"
  port             = 443
  scheme           = "https"

  custom_certificate = {
    certificate = file("certs/legacy-fullchain.pem")
    private_key = file("certs/legacy-key.pem")
  }
}
```

## Attributes

| Attribute            | Type        | Required | Description |
|----------------------|-------------|----------|-------------|
| `cluster`            | String      | **Yes**  | UUID of the ingress cluster. Forces replacement. |
| `server_name`        | String      | **Yes**  | Fully qualified hostname, e.g. `shop.example.com`. |
| `port`               | Number      | **Yes**  | Upstream port (1–65535). |
| `target_kind`        | String      | No       | `vm`, `kubernetes` or `address`. Inferred when omitted. |
| `virtualmachine`     | String      | No       | Upstream VM UUID (`vm`). |
| `interface`          | String      | No       | Upstream VM interface UUID (`vm`). |
| `kubernetes_cluster` | String      | No       | Upstream Kubernetes cluster UUID (`kubernetes`). |
| `upstream_address`   | String      | No       | Upstream IP. Required for `address`, optional for `kubernetes`, derived for `vm`. |
| `scheme`             | String      | No       | `http` (default) or `https`. |
| `path`               | String      | No       | Path served by the route. Default `/`. |
| `ssl`                | Bool        | No       | Serve over HTTPS. Default `true`. |
| `settings`           | Map(String) | No       | BunkerWeb setting overrides. |
| `secret_settings`    | Map(String) | No       | Write-only BunkerWeb settings. Sensitive. |
| `locations`          | List(Object)| No       | Extra locations, see below. |
| `waf_exceptions`     | List(Object)| No       | WAF exceptions, see below. |
| `custom_certificate` | Object      | No       | Custom certificate, see below. |

**Read-only attributes:** `id` (String), `upstream` (String), `secrets_set` (List of String).

`locations` items:

| Attribute            | Type        | Required | Description |
|----------------------|-------------|----------|-------------|
| `path`               | String      | **Yes**  | Location path; must start with `/` and not be `/`. |
| `port`               | Number      | **Yes**  | Upstream port. |
| `target_kind`        | String      | No       | `vm`, `kubernetes` or `address`. Inferred when omitted. |
| `virtualmachine`     | String      | No       | Upstream VM UUID. |
| `interface`          | String      | No       | Upstream VM interface UUID. |
| `kubernetes_cluster` | String      | No       | Upstream Kubernetes cluster UUID. |
| `upstream_address`   | String      | No       | Upstream IP address. |
| `scheme`             | String      | No       | `http` (default) or `https`. |
| `settings`           | Map(String) | No       | Location-level BunkerWeb settings. |

Read-only: `id` (String), `slot` (Number), `upstream` (String).

`waf_exceptions` items:

| Attribute    | Type   | Required | Description |
|--------------|--------|----------|-------------|
| `reason`     | String | **Yes**  | Why the exception is needed. |
| `kind`       | String | No       | `structured` (default) or `raw`. |
| `rules`      | String | No       | ModSecurity rule IDs to disable (structured). |
| `path`       | String | No       | Path the exception applies to (structured). |
| `path_match` | String | No       | `exact` (default) or `prefix` (structured). |
| `target`     | String | No       | Request part to exclude, e.g. `ARGS:q` (structured). |
| `raw`        | String | No       | Raw ModSecurity configuration (raw). |

Read-only: `id` (Number).

`custom_certificate`:

| Attribute     | Type   | Required | Description |
|---------------|--------|----------|-------------|
| `certificate` | String | **Yes**  | PEM certificate with intermediates, or base64 PFX/P7B bundle. Sensitive. |
| `format`      | String | No       | `PEM` (default), `PFX` or `P7B`. |
| `private_key` | String | No       | PEM private key (PEM and P7B). Sensitive. |
| `passphrase`  | String | No       | Passphrase of the PFX bundle or encrypted key. Sensitive. |

Read-only: `id`, `common_name`, `subject_alt_names` (List of String), `issuer`, `serial_number`, `fingerprint_sha256`, `valid_from_timestamp`, `valid_to_timestamp`.


## Import

The API never returns `secret_settings` or the `custom_certificate` inputs, so the first `terraform apply` after an import sends them again: secret settings are re-set and a configured certificate is re-uploaded.

```shell
terraform import mcs_secureingress_route.shop <route-uuid>
```
