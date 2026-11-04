data "mcs_public_ip_address" "web" {
  ip_address = "203.0.113.10"
}

output "web_ip_status" {
  value = data.mcs_public_ip_address.web.status
}
