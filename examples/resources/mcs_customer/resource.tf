data "mcs_contact" "admin" {
  name = "Example Corp"
}

resource "mcs_customer" "example" {
  name           = "Example Customer"
  tech_contacts  = [data.mcs_contact.admin.id]
  admin_contacts = [data.mcs_contact.admin.id]
}
