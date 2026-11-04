resource "mcs_certificate" "web_cert" {
  name         = "web-ssl-cert"
  ca           = false
  loadbalancer = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
