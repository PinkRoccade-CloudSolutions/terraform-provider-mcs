data "mcs_firewall" "internet" {
  name = "Gemnet"
}

output "firewall_id" {
  value = data.mcs_firewall.internet.id
}
