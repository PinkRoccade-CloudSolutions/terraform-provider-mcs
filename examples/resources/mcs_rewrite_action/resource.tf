data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_rewrite_action" "replace_host" {
  name              = "replace-host"
  type              = "replace"
  target            = "HTTP.REQ.HOSTNAME"
  stringbuilderexpr = "\"example.org\""
  search            = "example.com"
  comment           = "Replace host header"
  customer          = data.mcs_customer.example.id
  loadbalancer      = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
