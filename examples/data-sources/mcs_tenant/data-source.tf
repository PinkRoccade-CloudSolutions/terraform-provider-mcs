data "mcs_tenant" "current" {}

output "tenant_name" {
  value = data.mcs_tenant.current.name
}
