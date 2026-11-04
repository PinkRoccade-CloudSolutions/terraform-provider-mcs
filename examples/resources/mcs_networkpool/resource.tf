resource "mcs_networkpool" "lan" {
  name        = "Production LAN Pool"
  network     = "10.0.0.0/8"
  description = "Internal LAN networks"
  type        = "lan"
  enabled     = true
}
