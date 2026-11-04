---
page_title: "mcs_firewall_rule Resource - mcs"
subcategory: "Firewall"
description: |-
  Manages a firewall policy rule.
---

# mcs_firewall_rule (Resource)

Manages a firewall policy rule.

## Example Usage

```hcl
data "mcs_network" "production" {
  name = "001.2002.USR"
}

resource "mcs_firewall_rule" "allow_https" {
  domain   = "VDOM-001"
  enabled  = true
  src      = ["client-net"]
  dst      = ["web-servers-group"]
  src_intf = [data.mcs_network.production.name]
  dst_intf = ["any"]
  service  = ["HTTPS"]
  action   = true
  comment  = "Allow HTTPS to web servers"
}
```

## Attributes

| Attribute | Type         | Required | Description |
|----------|-------------|----------|-------------|
| `domain` | String       | **Yes**  | Firewall domain (VDOM). Changing this forces a new resource. |
| `enabled` | Bool        | **Yes**  | Whether the rule is enabled. |
| `action` | Bool         | **Yes**  | `true` = allow, `false` = deny. |
| `src`    | List(String) | No       | Source address objects/groups. Removing the attribute clears it. |
| `dst`    | List(String) | No       | Destination address objects/groups. Removing the attribute clears it. |
| `src_intf` | List(String) | No     | Source interfaces. Removing the attribute clears it. |
| `dst_intf` | List(String) | No     | Destination interfaces. Removing the attribute clears it. |
| `service` | List(String) | No      | Service objects/groups (e.g. `"HTTPS"`, `"SSH"`). Removing the attribute clears it. |
| `comment` | String      | No       | Comment / description. Server value is kept when omitted. |

**Read-only attributes:**

| Attribute           | Type         | Description |
|--------------------|-------------|-------------|
| `id`               | String       | Rule ID (same as `policyid`). |
| `uuid`             | String       | Firewall UUID. |
| `policyid`         | Number       | Policy ID (used internally for API operations). |
| `group`            | String       | Rule group. |
| `origin`           | String       | Origin of the rule. |
| `used`             | Bool         | Whether the rule is in use. |
| `compliant`        | Bool         | Whether the rule is compliant. |
| `hit_count`        | Number       | Number of hits on the rule. |
| `last_hit`         | String       | Timestamp of the last hit. |
| `compliancy_errors` | List(String) | Compliancy errors reported for the rule. |

## Import

Import by firewall domain and numeric policy ID, separated by `/`:

```shell
terraform import mcs_firewall_rule.allow_https <domain>/<policyid>
```
