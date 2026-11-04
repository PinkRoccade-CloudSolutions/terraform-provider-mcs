resource "mcs_firewall_service" "custom_app" {
  domain        = "VDOM-PROD"
  name          = "APP-8443"
  protocol      = "TCP/UDP/SCTP"
  tcp_portrange = ["8443"]
  comment       = "Custom application port"
}
