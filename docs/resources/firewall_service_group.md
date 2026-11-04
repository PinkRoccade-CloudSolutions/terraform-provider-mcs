---
page_title: "mcs_firewall_service_group Resource - mcs"
subcategory: "Firewall"
description: |-
  Manages a group of firewall services.
---

# mcs_firewall_service_group (Resource)

Manages a group of firewall services.

## Example Usage

```hcl
resource "mcs_firewall_service_group" "web_services" {
  domain  = "VDOM-PROD"
  name    = "web-services"
  comment = "HTTP and HTTPS services"
  member  = ["HTTP", "HTTPS", "custom-app"]
}
```

## Attributes

| Attribute | Type         | Required | Description |
|----------|-------------|----------|-------------|
| `domain` | String       | **Yes**  | Firewall domain (VDOM). Changing this forces a new resource. |
| `name`   | String       | **Yes**  | Service group name. Renaming updates in place. |
| `comment` | String      | No       | Comment / description. Server value is kept when omitted. |
| `member` | List(String) | No       | List of member service names. Removing the attribute clears the members. |

**Read-only attributes:** `id` (String), `uuid` (String), `used` (Bool).

## Import

Import by firewall domain and name, separated by `/`:

```shell
terraform import mcs_firewall_service_group.web_services <domain>/<name>
```
