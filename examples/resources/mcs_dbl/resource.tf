resource "mcs_dbl" "blocked_ip" {
  ipaddress  = "192.0.2.100"
  source     = "manual"
  persistent = true
  blackholed = false
  itsm       = "INC0012345"
}
