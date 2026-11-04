data "mcs_public_ip_address" "web" {
  ip_address = "203.0.113.10"
}

data "mcs_cs_policy" "routing" {
  name = "route-by-url"
}

data "mcs_certificate" "web_cert" {
  name = "web-cert"
}

data "mcs_certificate" "client_ca" {
  name = "client-ca"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_csv_server" "web_frontend" {
  name           = "web-csvserver"
  ufname         = "web-frontend"
  type           = "ssl"
  port           = 443
  ipaddress      = data.mcs_public_ip_address.web.id
  policies       = [data.mcs_cs_policy.routing.id]
  certificate    = [data.mcs_certificate.web_cert.id]
  ca_certificate = [data.mcs_certificate.client_ca.id]
  clientauth     = true
  clientcert     = "Mandatory"
  customer       = data.mcs_customer.example.id
  loadbalancer   = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
