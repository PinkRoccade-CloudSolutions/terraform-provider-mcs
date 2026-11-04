---
page_title: "mcs_job Data Source - mcs"
subcategory: "Compute"
description: |-
  Look up a specific job by ID. This data source always requires an `id`.
---

# mcs_job (Data Source)

Look up a specific job by ID. This data source always requires an `id`.

## Example Usage

```hcl
data "mcs_job" "deployment" {
  id = 12345
}

output "job_status" {
  value = "${data.mcs_job.deployment.status}: ${data.mcs_job.deployment.result}"
}
```

## Attributes

| Attribute            | Type   | Mode     | Description |
|---------------------|--------|----------|-------------|
| `id`                | Number | **Required** | Job ID. |
| `jobname`           | String | Computed | Name of the job. |
| `status`            | String | Computed | Job status name (flattened from the API's nested `status` object). |
| `result`            | String | Computed | Job result name (flattened from the API's nested `result` object). |
| `timestamp`         | String | Computed | Job start timestamp. |
| `endtime`           | String | Computed | Job end timestamp; null while the job is still running. |
| `message`           | String | Computed | Job status message. |
| `dryrun`            | Bool   | Computed | Whether this was a dry run. |
| `continue_on_failure` | Bool | Computed | Whether the job continues on failure. |
| `can_force_continue_transit` | String | Computed | Whether the job can be forced to continue a transit step. |
| `parent`            | Number | Computed | ID of the parent job, if chained. |
| `sub_jobs`          | List of Number | Computed | IDs of the sub jobs. |
| `created_at_timestamp` | String | Computed | Creation time. |
| `updated_at_timestamp` | String | Computed | Last update time. |
| `created_by_user`   | Number | Computed | ID of the user who created the job (may be null). |
| `updated_by_user`   | Number | Computed | ID of the user who last updated the job (may be null). |
