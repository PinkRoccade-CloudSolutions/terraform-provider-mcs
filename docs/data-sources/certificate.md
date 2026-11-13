---
page_title: "mcs_certificate Data Source - mcs"
subcategory: "Load Balancing"
description: |-
  Look up load balancer certificates. Provide `name` or `id` for a single match, or omit both to list all.
---

# mcs_certificate (Data Source)

Look up load balancer certificates. Provide `name` or `id` for a single match, or omit both to list all.

## Example Usage

```hcl
data "mcs_certificate" "web_cert" {
  name = "web-ssl-cert"
}
```

## Attributes

| Attribute            | Type   | Mode     | Description |
|---------------------|--------|----------|-------------|
| `name`              | String | Optional | Exact certificate name. |
| `id`                | String | Optional/Computed | Certificate UUID. |
| `ca`                | Bool   | Computed | Whether this is a CA certificate. |
| `valid_to_timestamp` | String | Computed | Certificate expiry timestamp. |
| `loadbalancer`      | String | Computed | UUID of the load balancer. |
| `customer`          | String | Computed | Customer identifier. |
| `protected`         | Bool   | Computed | Whether the certificate is protected from changes. |
| `subject_alt_names` | List of String | Computed | DNS names from the Subject Alternative Name extension. |
| `subject_alt_ips`   | List of String | Computed | IP addresses from the Subject Alternative Name extension. |
| `endpoints`         | List   | Computed | Published endpoints (domain names) that use this certificate or that its SANs cover; see below. |
| `certificates`      | List   | Computed | All certificates (populated when neither `name` nor `id` is set). |

**Nested `certificates` attributes:** `id`, `name`, `valid_to_timestamp`, `loadbalancer`, `customer` (String); `ca`, `protected` (Bool); `subject_alt_names`, `subject_alt_ips` (List of String); `endpoints` (List, as below) — all Computed.

**Nested `endpoints` attributes** (all Computed):

| Attribute        | Type   | Description |
|------------------|--------|-------------|
| `kind`           | String | `http` (hostname route on a content switch) or `l4` (load balancing vserver). |
| `hostname`       | String | Hostname of the endpoint; empty for an L4 endpoint or a vserver without hostname routing. |
| `policy_id`      | String | UUID of the content switching policy routing the hostname; null for L4. |
| `policy_name`    | String | Name of that policy. |
| `csvserver_id`   | String | UUID of the content switching vserver; null for L4. |
| `csvserver_name` | String | Name of that vserver. |
| `lbvserver_id`   | String | UUID of the load balancing vserver. |
| `lbvserver_name` | String | Name of that vserver. |
| `vip`            | String | Virtual IP address of the endpoint. |
| `port`           | Number | Port of the endpoint. |
| `bound`          | Bool   | Whether the certificate is bound to the vserver serving this endpoint. |
| `covered`        | Bool   | Whether the certificate's SANs cover the hostname; null when unknown (no SANs or no hostname). |

```hcl
data "mcs_certificate" "web_cert" {
  name = "web-ssl-cert"
}

output "unbound_hostnames" {
  value = [for e in data.mcs_certificate.web_cert.endpoints : e.hostname if !e.bound]
}
```
