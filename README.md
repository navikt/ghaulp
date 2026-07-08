# Github Audit Log Poller

Polls the GitHub Audit Log API every 5 minutes and posts to Slack.

## How it works

1. Read the last-seen audit log cursor from `cursor.txt` in the GCS bucket.
2. Fetch new audit log entries from the GitHub API using the stored cursor.
3. Filter to `integration_installation`, `integration_installation_request` and `integration.update`.
4. Post each event to Slack (oldest first).
5. Write the newest cursor back to GCS.

## Event types

| Action | Description |
|---|---|
| `integration_installation` | A GitHub App was installed or uninstalled |
| `integration_installation_request` | A user requested installation of a GitHub App |
| `integration.update` | A user requested a permission change for an installed GitHub App |

### Permission change alerts (`integration.update`)

When a user requests updated permissions for a GitHub App, the Slack message includes a direct link to the GitHub permissions approval page so admins can review and accept the change:

```
*integration.update* by `octocat` — app: `kitten-playground` (2026-07-08T07:34:39+02:00)
<https://github.com/organizations/{org}/settings/installations/{id}/permissions/update|Review and approve permissions>
```

The installation ID is resolved at runtime via `GET /orgs/{org}/installations`, which requires the app to have **Read access to organization administration**. If the lookup fails the alert is still posted without the link.
