data "mcs_application" "webshop" {
  name = "webshop"
}

data "mcs_cs_action" "webshop" {
  name = "route-to-webshop"
}

resource "mcs_cs_policy" "webshop" {
  name        = "route-webshop"
  action      = data.mcs_cs_action.webshop.id
  expression  = "HTTP.REQ.HOSTNAME.EQ(\"shop.example.com\")"
  application = data.mcs_application.webshop.id
}
