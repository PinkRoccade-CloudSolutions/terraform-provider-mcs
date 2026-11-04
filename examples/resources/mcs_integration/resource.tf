variable "prtg_password" {
  type      = string
  sensitive = true
}

resource "mcs_integration" "prtg" {
  name        = "prtg-prod"
  catalog_key = "prtg"
  baseurl     = "https://prtg.example.com"
  username    = "svc-mcs"
  secret      = var.prtg_password
}
