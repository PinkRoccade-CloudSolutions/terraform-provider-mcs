<!-- TOC entry (under "Firewall", before mcs_firewall_object): -->
<!--     - [mcs_firewall (Resource)](#mcs_firewall-resource) -->

#### mcs_firewall (Resource)

Manages a firewall (`/api/networking/firewalls/`). Updates use PATCH.

##### Example

```hcl
resource "mcs_firewall" "edge" {
  customer            = mcs_customer.example.id
  device              = "fmg-01"
  name                = "edge-fw"
  type                = "internet"
  context             = "VDOM-PROD"
  external_interface  = "port1"
  internal_interface  = "port2"
  nat_ip_sync_enabled = false
}
```

##### Attributes

| Attribute                 | Type   | Required | Description |
|--------------------------|--------|----------|-------------|
| `customer`               | String | **Yes**  | Customer the firewall belongs to. Changing this forces a new resource. |
| `device`                 | String | **Yes**  | Firewall device (see `/api/networking/firewalls/device-options/`). Changing this forces a new resource. |
| `name`                   | String | No       | Firewall name. |
| `description`            | String | No       | Firewall description. |
| `type`                   | String | No       | `internet` or `wan`. |
| `context`                | String | No       | VDOM on Fortimanager or Device Group on Panorama. |
| `external_interface`     | String | No       | Name of the external (internet or WAN facing) interface. |
| `internal_interface`     | String | No       | Name of the internal (VDOM or transit facing) interface. |
| `default_log_profile`    | String | No       | Default log profile used when creating firewall rules. |
| `default_protect_profile` | String | No      | Default protect group profile used when creating firewall rules. |
| `multi_tenant`           | Bool   | No       | Whether the device is multi tenant; `tag_name` is then used for rule separation. |
| `tag_name`               | String | No       | TAG name for object lookups on a multi tenant firewall (PaloAlto only). |
| `nat_ip_sync_enabled`    | Bool   | No       | Opt in to the daily Panorama NAT public IP import job for this device group. |

Optional attributes that are not set take the value assigned by the API.

**Read-only attributes:** `id` (String — firewall UUID), `customer_name` (String), `device_name` (String), `platform` (String), `supports_threat_protection` (Bool).

##### Import

```shell
terraform import mcs_firewall.edge <firewall-uuid>
```
