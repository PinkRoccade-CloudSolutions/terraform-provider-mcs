data "mcs_interface" "eth0" {
  name = "eth0"
}

output "interface_ip" {
  value = data.mcs_interface.eth0.ipaddress
}
