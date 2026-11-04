resource "mcs_firewall_object_group" "web_servers" {
  domain  = "VDOM-001"
  name    = "web-servers-group"
  comment = "All web server objects"
  member  = ["web-server-01"]
}
