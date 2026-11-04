resource "mcs_firewall_service_group" "web_services" {
  domain  = "VDOM-PROD"
  name    = "web-services"
  comment = "HTTP and HTTPS services"
  member  = ["HTTP", "HTTPS", "custom-app"]
}
