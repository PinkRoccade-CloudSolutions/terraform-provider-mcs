data "mcs_rewrite_action" "replace_host" {
  name = "replace-host"
}

data "mcs_customer" "example" {
  name = "Example Corp"
}

resource "mcs_rewrite_policy" "rewrite_requests" {
  name                   = "rewrite-requests"
  rule                   = "HTTP.REQ.URL.CONTAINS(\"/old\")"
  action                 = data.mcs_rewrite_action.replace_host.id
  undefaction            = "NOREWRITE"
  comment                = "Rewrite incoming requests"
  priority               = 100
  bindpoint              = "REQUEST"
  gotopriorityexpression = "NEXT"
  customer               = data.mcs_customer.example.id
  loadbalancer           = "6b68cf7b-6d6e-459d-a583-b02fdccfadbd"
}
