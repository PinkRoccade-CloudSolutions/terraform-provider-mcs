variable "shop_basic_auth_password" {
  type      = string
  sensitive = true
}

data "mcs_ingress_cluster" "edge" {
  name = "edge"
}

data "mcs_virtualmachine" "web" {
  name = "web-server-01"
}

data "mcs_interface" "web" {
  name = "eth0"
}

resource "mcs_secureingress_route" "shop" {
  cluster        = data.mcs_ingress_cluster.edge.id
  server_name    = "shop.example.com"
  virtualmachine = data.mcs_virtualmachine.web.id
  interface      = data.mcs_interface.web.id
  port           = 8080

  settings = {
    USE_MODSECURITY = "yes"
    ALLOWED_METHODS = "GET|POST"
    USE_AUTH_BASIC  = "yes"
  }
  secret_settings = {
    AUTH_BASIC_PASSWORD = var.shop_basic_auth_password
  }

  locations = [{
    path             = "/api"
    upstream_address = "10.0.0.9"
    port             = 9000
  }]

  waf_exceptions = [{
    rules  = "942100"
    path   = "/search"
    target = "ARGS:q"
    reason = "Search box triggers SQLi false positives"
  }]
}
