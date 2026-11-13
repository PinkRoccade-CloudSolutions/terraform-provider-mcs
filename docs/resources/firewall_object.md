---
page_title: "mcs_firewall_object Resource - mcs"
subcategory: "Firewall"
description: |-
  Manages a firewall address object.
---

# mcs_firewall_object (Resource)

Manages a firewall address object.

-> **Note:** Changes to this resource are staged on the firewall domain and only go live after a commit. Add an [`mcs_firewall_commit`](firewall_commit.md) for the domain to commit them as part of `terraform apply`.

## Example Usage

```hcl
resource "mcs_firewall_object" "web_server" {
  domain  = "VDOM-001"
  name    = "web-server-prod"
  address = "10.0.1.10"
  subnet  = "255.255.255.255"
  comment = "Production web server"
}
```

## Attributes

| Attribute | Type   | Required | Description |
|----------|--------|----------|-------------|
| `domain` | String | **Yes**  | Firewall domain (VDOM). Changing this forces a new resource. |
| `name`   | String | **Yes**  | Object name. Renaming updates in place. |
| `address` | String | **Yes** | IP address. |
| `subnet` | String | No       | Subnet mask, e.g. `255.255.255.255`. Leave unset for an object without a mask; removing it clears the mask. Cannot be empty. |
| `comment` | String | No      | Comment / description. Server value is kept when omitted. |

**Read-only attributes:** `id` (String), `uuid` (String — firewall UUID), `used` (Bool — whether the object is in use by a policy), `managed` (Bool — whether the object is managed by MCS).

## Import

Import by firewall domain and name, separated by `/`:

```shell
terraform import mcs_firewall_object.web_server <domain>/<name>
```
