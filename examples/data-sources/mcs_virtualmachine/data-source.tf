data "mcs_virtualmachine" "webserver" {
  name = "web-server-01"
}

output "vm_cpu" {
  value = data.mcs_virtualmachine.webserver.cpu
}

output "vm_primary_ip" {
  value = data.mcs_virtualmachine.webserver.interfaces[0].ipaddress
}
