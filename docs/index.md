---
page_title: "Provider: MCS"
description: |-
  The MCS provider manages Mijn Cloud Solutions (MCS) resources: networking, firewalls, load balancing, secure ingress, DNS, VPN, monitoring and tenant administration.
---

# MCS Provider

The MCS (Mijn Cloud Solutions) Terraform provider allows you to manage cloud infrastructure resources through the MCS API. It supports managing networking, firewalls, load balancing, secure ingress, DNS, virtual machines, VPN, monitoring, and tenant administration.

Every resource and data source has its own page in the navigation, with a usage example, the full argument and attribute reference and, for resources, import instructions. They are grouped by area:

- **Tenant & Customer Management:** contacts, customers, quota profiles, integrations
- **Networking:** networks, network and IP pools, public IP addresses, NAT translations, site-to-site VPNs, load balancer pools
- **Secure Ingress:** secure ingress firewalls, ingress clusters, routes
- **Firewall:** firewalls, objects, object groups, services, service groups, rules
- **Load Balancing:** applications, certificates, monitors, service groups, virtual servers, content switching and rewrite actions/policies
- **Monitoring, DNS and Deny/Block Lists:** IP monitors, DNS domains and entries, IP and domain block lists
- **Compute:** virtual machines, disks, interfaces, virtual datacenters and jobs (data sources only)

## Provider Configuration

```hcl
terraform {
  required_providers {
    mcs = {
      source  = "PinkRoccade-CloudSolutions/mcs"
      version = "~> 0.4"
    }
  }
}

variable "mcs_token" {
  type      = string
  sensitive = true
}

provider "mcs" {
  host     = "https://mcs.example.com"
  token    = var.mcs_token
  insecure = false
}
```

### Provider Arguments

| Attribute  | Type   | Required | Description |
|------------|--------|----------|-------------|
| `host`     | String | Yes*     | Base URL of the MCS API (e.g. `https://mcs.example.com`). Can also be set via the `MCS_HOST` environment variable. |
| `token`    | String | Yes*     | API token for MCS authentication. Can also be set via the `MCS_TOKEN` environment variable. This value is marked sensitive and will not appear in plan output. |
| `insecure` | Bool   | No       | Skip TLS certificate verification. Only use for development with self-signed certificates. Defaults to `false`. |

> **\* Note:** Both `host` and `token` are marked as optional in the schema to allow configuration via environment variables alone. However, at least one source (config or environment variable) must provide a value for each; otherwise the provider will return an error.

---

## Authentication

The provider authenticates to the MCS API using token-based authentication. You can supply credentials in two ways:

### Option 1: Provider block

```hcl
variable "mcs_token" {
  type      = string
  sensitive = true
}

provider "mcs" {
  host  = "https://mcs.example.com"
  token = var.mcs_token
}
```

### Option 2: Environment variables

```bash
export MCS_HOST="https://mcs.example.com"
export MCS_TOKEN="your-api-token-here"
```

```hcl
provider "mcs" {}
```

Environment variables are evaluated first, then overridden by explicit provider block values if set.

---

## Complete Example

The following example demonstrates a realistic workflow: setting up a customer, looking up infrastructure, allocating a public IP, and creating a NAT translation.

```hcl
terraform {
  required_providers {
    mcs = {
      source  = "PinkRoccade-CloudSolutions/mcs"
      version = "~> 0.4"
    }
  }
}

# Configure provider using environment variables or explicit values
variable "mcs_token" {
  type      = string
  sensitive = true
}

provider "mcs" {
  host  = "https://mcs.example.com"
  token = var.mcs_token
}

# --- Look up existing infrastructure ---

data "mcs_tenant" "current" {}

data "mcs_virtualmachine" "webserver" {
  name = "web-server-01"
}

data "mcs_firewall" "internet" {
  name = "Internet-FW"
}

data "mcs_ippool" "nat_pool" {
  name = "NAT Pool"
}

data "mcs_network" "production" {
  name = "001.2002.USR"
}

data "mcs_zone" "bc_zone" {
  name = "Business Critical - AM5"
}

# --- Create tenant resources ---

resource "mcs_contact" "admin" {
  company   = "Example Corp"
  firstname = "Jane"
  lastname  = "Smith"
  email     = "jane.smith@example.com"
  phone     = "+31 6 98765432"
}

resource "mcs_customer" "production" {
  name           = "Production Customer"
  tech_contacts  = [mcs_contact.admin.id]
  admin_contacts = [mcs_contact.admin.id]
}

# --- Networking ---

resource "mcs_public_ip_address" "web" {
  pool        = data.mcs_ippool.nat_pool.id
  description = "Web server public IP"
  type        = "nat"
  customer    = mcs_customer.production.id
}

resource "mcs_nat_translation" "web_nat" {
  public_ip        = mcs_public_ip_address.web.id
  interface        = data.mcs_virtualmachine.webserver.interfaces[0].id
  firewall         = data.mcs_firewall.internet.id
  translation_type = "one_to_one"
  customer         = mcs_customer.production.id
  description      = "1:1 NAT for web server"
}

# --- Firewall rules ---

resource "mcs_firewall_object" "web_server" {
  domain  = "VDOM-001"
  name    = "web-server-prod"
  address = data.mcs_virtualmachine.webserver.interfaces[0].ipaddress
  subnet  = "255.255.255.255"
  comment = "Production web server"
}

resource "mcs_firewall_rule" "allow_https" {
  domain   = "VDOM-001"
  enabled  = true
  src      = ["all"]
  dst      = [mcs_firewall_object.web_server.name]
  src_intf = ["any"]
  dst_intf = [data.mcs_network.production.name]
  service  = ["HTTPS"]
  action   = true
  comment  = "Allow HTTPS to web server"
}

# --- Outputs ---

output "public_ip" {
  value = mcs_public_ip_address.web.ip_address
}

output "vm_details" {
  value = {
    name   = data.mcs_virtualmachine.webserver.name
    cpu    = data.mcs_virtualmachine.webserver.cpu
    memory = data.mcs_virtualmachine.webserver.memory
    os     = data.mcs_virtualmachine.webserver.os
  }
}
```
