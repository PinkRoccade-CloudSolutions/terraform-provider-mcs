resource "mcs_application" "webshop" {
  name             = "webshop"
  pentested        = true
  pentest_type     = "graybox"
  pentest_date     = "2026-03-15"
  pentest_findings = 4
}
