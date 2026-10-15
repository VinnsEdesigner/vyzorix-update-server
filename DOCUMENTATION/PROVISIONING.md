# Provisioning

Provisioning lets you declare what a server should contain in a single file, and
have the server create it at boot. Organizations, operators, groups, API keys,
permission grants, contact points, alert rules, service accounts, annotations —
all of it, from one YAML or JSON document.

It's how you go from "a fresh database" to "a working install" without clicking
through the dashboard. Point `PROVISIONING_FILE` at a file and restart; the
server reads it once and creates what's missing.

---

## How it runs

At startup, if `PROVISIONING_FILE` is set, the server loads the file, applies
each section in a fixed order, and logs what it did. A missing file is not an
error — the server logs that it found nothing and continues.

The order matters, because later sections reference earlier ones by name:

```
organizations
  → operators
  → device_groups
  → api_keys
  → permission_grants
  → contact_points
  → alert_rules
  → service_accounts
  → annotations
```

Operators are created **before** organizations on purpose: creating an org
inserts a membership and records who created it, both of which point at an
operator. An organization with no operator to own it can't exist, so the
provisioner picks a creator from the operators you declared — the first one whose
`org_name` matches the organization, or the first operator if none matches. If
there is no operator at all, that org is skipped with a clear log line rather
than created in a broken state.

---

## It is idempotent

Run it on every boot. Every resource is looked up before it's created:

- Organizations, operators, groups, alert rules, contact points, service
  accounts, and annotations are matched by name (or email) within their
  organization and **skipped** if they already exist.
- API keys are matched by name; a key that exists is skipped.

Nothing is duplicated and nothing is overwritten. A second boot against the same
database does nothing but log "already exists" lines. That means the file is safe
to keep in your deployment config and let it run on every start.

---

## The file

YAML or JSON — the server detects which by looking at the content. Every section
is optional; declare only what you need.

```yaml
organizations:
  - name: Default Org
    description: Auto-provisioned default organization

operators:
  - email: admin@vyzorix.local
    name: Admin
    password: ChangeMe123!
    org_name: Default Org
    role: super_admin

device_groups:
  - name: production
    org_name: Default Org
    operator_emails: [admin@vyzorix.local]

api_keys:
  - name: ci-pipeline
    org_name: Default Org
    scope: admin

permission_grants:
  - subject_type: operator
    subject_id: admin@vyzorix.local
    org_name: Default Org
    scope: "devices:*"
    actions: [device.manage]

contact_points:
  - name: oncall-email
    org_name: Default Org
    channel: email
    enabled: true
    config: { to: oncall@vyzorix.local }

alert_rules:
  - name: Fleet offline
    org_name: Default Org
    metric: device_offline_count
    condition: gt
    threshold: 10
    for_seconds: 300
    notify_interval_seconds: 3600
    enabled: true

service_accounts:
  - name: nightly-report
    org_name: Default Org

annotations:
  - title: Initial provisioning
    org_name: Default Org
    text: Created by the boot provisioning file.
    source: provisioning
    tags: [bootstrap]
```

See `provisioning.example.yaml` and `provisioning.example.json` in the repo root
for the full shape of every section.

### Section reference

| Section | Fields |
|:--------|:-------|
| `organizations` | `name`, `description` |
| `operators` | `email`, `name`, `password`, `org_name`, `role` (`super_admin` or `admin`) |
| `device_groups` | `name`, `org_name`, `operator_emails[]` |
| `api_keys` | `name`, `org_name`, `scope` (`read` / `write` / `admin`) |
| `permission_grants` | `subject_type`, `subject_id`, `org_name`, `scope`, `actions[]` |
| `contact_points` | `name`, `org_name`, `channel` (`email` / `webhook` / `slack`), `enabled`, `config{}` |
| `alert_rules` | `name`, `org_name`, `metric`, `condition`, `threshold`, `for_seconds`, `notify_interval_seconds`, `webhook_url`, `enabled` |
| `service_accounts` | `name`, `org_name` |
| `annotations` | `title`, `org_name`, `text`, `source`, `start_time`, `end_time`, `tags[]` |

`config` for a contact point carries the channel's keys: `to` for email, `url`
for webhook, `webhook` for Slack. The same validation the API applies runs here,
so a malformed contact point fails at boot, not silently.

---

## Configuration

| Variable | Meaning |
|:---------|:--------|
| `PROVISIONING_FILE` | Path to the provisioning file. Unset means "don't provision". |

---

## What it's for

- **First boot.** A new deployment gets its default org, admin, and a CI key
  without any manual setup.
- **Repeatable environments.** The same file produces the same starting state in
  staging and production.
- **Declarative ops.** Alert rules and contact points live in version control
  next to the deployment config, reviewed like any other change.

Because it only creates what's missing, it's a *seed*, not a reconciliation loop:
it won't delete a resource you removed from the file, and it won't overwrite one
you edited in the dashboard. If you want something gone, remove it through the
API.
