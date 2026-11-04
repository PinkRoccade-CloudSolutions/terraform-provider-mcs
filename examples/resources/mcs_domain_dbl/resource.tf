resource "mcs_domain_dbl" "blocked_domain" {
  domainname = "malicious-site.example"
  source     = "manual"
  persistent = true
  labels     = ["malware"]
}
