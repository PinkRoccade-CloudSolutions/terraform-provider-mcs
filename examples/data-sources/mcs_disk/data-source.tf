data "mcs_disk" "primary" {
  name = "sda"
}

output "disk_size" {
  value = data.mcs_disk.primary.size
}
