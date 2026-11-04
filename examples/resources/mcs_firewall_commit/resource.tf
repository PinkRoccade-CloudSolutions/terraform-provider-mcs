locals {
  domain = "VDOM-001"
}

resource "mcs_firewall_object" "web" {
  domain  = local.domain
  name    = "web-servers"
  address = "10.0.10.0"
  subnet  = "255.255.255.0"
}

resource "mcs_firewall_rule" "allow_https" {
  domain   = local.domain
  enabled  = true
  src      = ["all"]
  dst      = [mcs_firewall_object.web.name]
  src_intf = ["any"]
  dst_intf = ["any"]
  service  = ["HTTPS"]
  action   = true
}

# Commits the domain once the rule and object have been applied, and again
# whenever either of them changes.
resource "mcs_firewall_commit" "vdom_001" {
  domain   = local.domain
  changeid = "CHG0012345"

  triggers = {
    firewall = sha1(jsonencode([
      mcs_firewall_object.web,
      mcs_firewall_rule.allow_https,
    ]))
  }

  depends_on = [
    mcs_firewall_object.web,
    mcs_firewall_rule.allow_https,
  ]
}
