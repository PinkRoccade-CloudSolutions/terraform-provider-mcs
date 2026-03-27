terraform {
  required_providers {
    mcs = {
      source = "registry.terraform.io/pinkroccade/mcs"
    }
  }
}

provider "mcs" {
  host  = "https://mcs.example.com"
  token = var.mcs_token
}

variable "mcs_token" {
  type      = string
  sensitive = true
}

# --- Data Sources ---

data "mcs_network" "all" {}

data "mcs_zone" "all" {}

data "mcs_domain" "all" {}

# --- Tenant Resources ---

resource "mcs_customer" "example" {
  name     = "Example Customer"
  tenant   = 1
}

resource "mcs_contact" "admin" {
  company   = "Example Corp"
  firstname = "John"
  lastname  = "Doe"
  email     = "john.doe@example.com"
  phone     = "+31 6 12345678"
  tenant    = 1
}

# --- Virtual Datacenter ---

resource "mcs_virtual_datacenter" "production" {
  name     = "production-vdc"
  customer = mcs_customer.example.id
}

# --- Load Balancing ---

resource "mcs_lb_servicegroup" "web_backend" {
  name          = "web-backend-sg"
  type          = "HTTP"
  state         = "enable"
  healthmonitor = "YES"
}

resource "mcs_lb_servicegroup_member" "web1" {
  address    = "10.0.1.10"
  port       = 8080
  servername = "web-server-01"
  weight     = 100
}

resource "mcs_lbv_server" "web_lb" {
  name         = "web-lbvserver"
  ipv46        = "10.0.0.100"
  port         = 443
  type         = "ssl"
  servicegroup = [mcs_lb_servicegroup.web_backend.id]
}

resource "mcs_csv_server" "web_frontend" {
  name   = "web-csvserver"
  ufname = "web-frontend"
  type   = "ssl"
  port   = 443
}

# --- Firewall ---

resource "mcs_firewall_object" "web_server" {
  domain  = "VDOM-PROD"
  name    = "web-server-01"
  address = "10.0.1.10"
  comment = "Production web server"
}

resource "mcs_firewall_rule" "allow_web" {
  domain  = "VDOM-PROD"
  enabled = true
  src     = ["any"]
  dst     = [mcs_firewall_object.web_server.name]
  service = ["HTTPS"]
  action  = true
  comment = "Allow HTTPS to web server"
}

resource "mcs_firewall_service" "custom_app" {
  domain       = "VDOM-PROD"
  name         = "APP-8443"
  protocol     = "TCP/UDP/SCTP"
  tcp_portrange = ["8443"]
  comment      = "Custom application port"
}

# --- VPN ---

resource "mcs_site_to_site_vpn" "office" {
  name  = "office-vpn-tunnel"
  state = "up"
}

# --- DBL ---

resource "mcs_dbl" "blocked_ip" {
  ipaddress  = "192.0.2.100"
  source     = "manual"
  persistent = true
}
