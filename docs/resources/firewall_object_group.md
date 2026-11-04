---
page_title: "mcs_firewall_object_group Resource - mcs"
subcategory: "Firewall"
description: |-
  Manages a firewall address group (a collection of firewall objects).
---

# mcs_firewall_object_group (Resource)

Manages a firewall address group (a collection of firewall objects).

## Example Usage

```hcl
resource "mcs_firewall_object_group" "web_servers" {
  domain  = "VDOM-001"
  name    = "web-servers-group"
  comment = "All web server objects"
  member  = ["web-server-01"]
}
```

## Attributes

| Attribute | Type         | Required | Description |
|----------|-------------|----------|-------------|
| `domain` | String       | **Yes**  | Firewall domain (VDOM). Changing this forces a new resource. |
| `name`   | String       | **Yes**  | Group name. Renaming updates in place. |
| `comment` | String      | No       | Comment / description. Server value is kept when omitted. |
| `member` | List(String) | No       | List of member object names. Removing the attribute clears the members. |

**Read-only attributes:** `id` (String), `uuid` (String), `used` (Bool).

## Import

Import by firewall domain and name, separated by `/`:

```shell
terraform import mcs_firewall_object_group.web_servers <domain>/<name>
```
