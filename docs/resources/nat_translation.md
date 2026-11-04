---
page_title: "mcs_nat_translation Resource - mcs"
subcategory: "Networking"
description: |-
  Manages a NAT translation between a public IP and a private interface. Supports 1:1 NAT, port forwarding and SNAT.
---

# mcs_nat_translation (Resource)

Manages a NAT translation between a public IP and a private interface. Supports 1:1 NAT, port forwarding and SNAT.

## Example Usage — 1:1 NAT

```hcl
data "mcs_public_ip_address" "web" {
  ip_address = "203.0.113.10"
}

data "mcs_virtualmachine" "webserver" {
  name = "web-server-01"
}

data "mcs_firewall" "internet" {
  name = "Internet-FW"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_nat_translation" "web_nat" {
  public_ip        = data.mcs_public_ip_address.web.id
  interface        = data.mcs_virtualmachine.webserver.interfaces[0].id
  firewall         = data.mcs_firewall.internet.id
  translation_type = "one_to_one"
  customer         = data.mcs_customer.example.id
  description      = "1:1 NAT for web server"
}
```

## Example Usage — Port forward

```hcl
data "mcs_public_ip_address" "web" {
  ip_address = "203.0.113.10"
}

data "mcs_interface" "eth0" {
  name = "eth0"
}

data "mcs_firewall" "internet" {
  name = "Internet-FW"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_nat_translation" "https_forward" {
  public_ip        = data.mcs_public_ip_address.web.id
  interface        = data.mcs_interface.eth0.id
  firewall         = data.mcs_firewall.internet.id
  translation_type = "port_forward"
  public_port      = 443
  private_port     = 8443
  protocol         = "tcp"
  customer         = data.mcs_customer.example.id
  description      = "HTTPS port forward"
}
```

## Example Usage — SNAT

```hcl
data "mcs_interface" "eth0" {
  name = "eth0"
}

data "mcs_firewall" "internet" {
  name = "Internet-FW"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

data "mcs_public_ip_address" "web" {
  ip_address = "203.0.113.10"
}

resource "mcs_nat_translation" "outbound" {
  interface                 = data.mcs_interface.eth0.id
  firewall                  = data.mcs_firewall.internet.id
  translation_type          = "snat"
  customer                  = data.mcs_customer.example.id
  snat_source_addresses     = ["10.0.0.0/24"]
  snat_translated_addresses = [data.mcs_public_ip_address.web.ip_address]
}
```

## Attributes

| Attribute          | Type   | Required | Description |
|-------------------|--------|----------|-------------|
| `firewall`        | String | **Yes**  | UUID of the firewall. Changing it forces a new resource. |
| `translation_type` | String | **Yes** | Translation type: `one_to_one`, `port_forward`, or `snat`. |
| `customer`        | String | **Yes**  | Customer identifier. Changing it forces a new resource. |
| `public_ip`       | String | No       | UUID of the public IP address. Removing it sends `null`. |
| `interface`       | String | No       | UUID of the private interface. Removing it sends `null`. |
| `public_port`     | Number | No       | Public port, 0–4294967295 (required for port forwarding). |
| `private_port`    | Number | No       | Private port, 0–4294967295 (required for port forwarding). |
| `protocol`        | String | No       | Protocol: `tcp`, `udp`, or `""`. |
| `description`     | String | No       | Description of the NAT translation (max 255). |
| `enabled`         | Bool   | No       | Whether the NAT translation is enabled. Defaults to `true` on create. |
| `snat_source_addresses` | List(String) | No | Source IPs/subnets to apply SNAT to (`snat` only). Removing it clears the list. |
| `snat_translated_addresses` | List(String) | No | Addresses used as the SNAT translation pool (`snat` only). Removing it clears the list. |

**Read-only attributes:**

| Attribute    | Type   | Description |
|-------------|--------|-------------|
| `id`        | String | UUID of the NAT translation. |
| `private_ip` | String | Resolved private IP address. |
| `translation` | String | Human-readable translation summary. |
| `state`     | String | Sync state: `synced`, `unsynced`, `error`, or `deleted`. |

## Import

```shell
terraform import mcs_nat_translation.web_nat <nat-translation-uuid>
```
