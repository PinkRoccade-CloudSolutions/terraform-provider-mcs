data "mcs_network" "production" {
  name = "001.2002.USR"
}

resource "mcs_firewall_rule" "allow_https" {
  domain   = "VDOM-001"
  enabled  = true
  src      = ["client-net"]
  dst      = ["web-servers-group"]
  src_intf = [data.mcs_network.production.name]
  dst_intf = ["any"]
  service  = ["HTTPS"]
  action   = true
  comment  = "Allow HTTPS to web servers"
}
