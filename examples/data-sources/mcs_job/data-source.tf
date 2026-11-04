data "mcs_job" "deployment" {
  id = 12345
}

output "job_status" {
  value = "${data.mcs_job.deployment.status}: ${data.mcs_job.deployment.result}"
}
