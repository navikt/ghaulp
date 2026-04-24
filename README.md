# Github Audit Log Poller

Polls the GitHub Audit Log API every 5 minutes and posts to Slack.

## How it works

1. Read the last-seen audit log cursor from `cursor.txt` in the GCS bucket.
2. Fetch new audit log entries from the GitHub API using the stored cursor.
3. Filter to `integration_installation` and `integration_installation_request`.
4. Post each event to Slack (oldest first).
5. Write the newest cursor back to GCS.
