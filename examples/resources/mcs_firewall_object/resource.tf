resource "mcs_firewall_object" "web_server" {
  domain  = "VDOM-001"
  name    = "web-server-prod"
  address = "10.0.1.10"
  subnet  = "255.255.255.255"
  comment = "Production web server"
}
