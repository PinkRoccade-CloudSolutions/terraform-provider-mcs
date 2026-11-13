---
page_title: "mcs_firewall_commit Resource - mcs"
subcategory: "Firewall"
description: |-
  Commits a domain's pending firewall changes so they go live.
---

# mcs_firewall_commit (Resource)

Commits a domain's pending firewall changes so they go live. Changes made with `mcs_firewall_rule`,
`mcs_firewall_object`, `mcs_firewall_object_group`, `mcs_firewall_service` and
`mcs_firewall_service_group` are staged until the domain is committed.

The resource picks the commit method from the domain's deploy flow:

- **V1** domains are committed through `POST /api/networking/domain/<domain>/commit/`.
- **V2** domains get a FortiManager install of their policy package. The provider waits for the
  install job to finish. An install pushes **everything** pending in the package, including
  changes made by other people.

Committing requires the `firewall.direct_management` permission; Change-Request-only users cannot
commit directly.

Terraform has no "run after everything" hook, so make the commit depend on the firewall resources
of the domain with `depends_on`, and put their values in `triggers`. Any change to `triggers`
replaces the resource, which commits again. When the API reports that nothing is pending (or an
install is already queued) the apply succeeds with a warning.

## Example Usage

```hcl
locals {
  domain = "VDOM-001"
}

resource "mcs_firewall_object" "web" {
  domain  = local.domain
  name    = "web-servers"
  address = "10.0.10.0"
  subnet  = "255.255.255.0"
}

resource "mcs_firewall_rule" "allow_https" {
  domain   = local.domain
  enabled  = true
  src      = ["all"]
  dst      = [mcs_firewall_object.web.name]
  src_intf = ["any"]
  dst_intf = ["any"]
  service  = ["HTTPS"]
  action   = true
}

# Commits the domain once the rule and object have been applied, and again
# whenever either of them changes.
resource "mcs_firewall_commit" "vdom_001" {
  domain   = local.domain
  changeid = "CHG0012345"

  triggers = {
    firewall = sha1(jsonencode([
      mcs_firewall_object.web,
      mcs_firewall_rule.allow_https,
    ]))
  }

  depends_on = [
    mcs_firewall_object.web,
    mcs_firewall_rule.allow_https,
  ]
}
```

## Attributes

| Attribute  | Type        | Required | Description |
|-----------|------------|----------|-------------|
| `domain`   | String      | **Yes**  | Firewall domain to commit as `VDOM-<vdom_id>` (e.g. `VDOM-001`), the same value as `mcs_firewall_rule.domain`. This is the domain's `vdom_id` with a `VDOM-` prefix, **not** its `name`; look it up with the `mcs_domain` data source if needed. Changing this commits the new domain. |
| `changeid` | String      | **Yes**  | Free-text change ID recorded with the commit, e.g. an ITSM ticket or pipeline run ID. Changing only this does not commit again. |
| `triggers` | Map(String) | No       | Arbitrary values; any change commits again. |
| `timeouts` | Block       | No       | `create` (default `20m`): how long to wait for a V2 FortiManager install. |

**Read-only attributes:**

| Attribute      | Type   | Description |
|---------------|-------|-------------|
| `id`           | String | `<domain>/<committed_at>` of the last commit. |
| `deploy_flow`  | String | Deploy flow of the domain at commit time: `v1` or `v2`. |
| `committed_at` | String | RFC 3339 time of the last commit. |
| `job_id`       | Number | V2 only: the FortiManager install job. Null for V1 or when nothing was pending. |

Destroying this resource only removes it from state; committed changes stay live.
