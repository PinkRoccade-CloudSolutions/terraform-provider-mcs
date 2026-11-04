---
page_title: "mcs_firewall_service Resource - mcs"
subcategory: "Firewall"
description: |-
  Manages a custom firewall service definition.
---

# mcs_firewall_service (Resource)

Manages a custom firewall service definition.

## Example Usage

```hcl
resource "mcs_firewall_service" "custom_app" {
  domain        = "VDOM-PROD"
  name          = "APP-8443"
  protocol      = "TCP/UDP/SCTP"
  tcp_portrange = ["8443"]
  comment       = "Custom application port"
}
```

## Attributes

| Attribute       | Type         | Required | Description |
|----------------|-------------|----------|-------------|
| `domain`       | String       | **Yes**  | Firewall domain (VDOM). Changing this forces a new resource. |
| `name`         | String       | **Yes**  | Service name. Renaming updates in place. |
| `protocol`     | String       | **Yes**  | Protocol type (e.g. `TCP/UDP/SCTP`). |
| `tcp_portrange` | List(String) | No      | TCP port ranges (e.g. `["80", "443", "8000-8999"]`). Removing the attribute clears it. |
| `udp_portrange` | List(String) | No      | UDP port ranges. Removing the attribute clears it. |
| `comment`      | String       | No       | Comment / description. Server value is kept when omitted. |

**Read-only attributes:** `id` (String), `uuid` (String), `used` (Bool).

## Import

Import by firewall domain and name, separated by `/`:

```shell
terraform import mcs_firewall_service.custom_app <domain>/<name>
```
