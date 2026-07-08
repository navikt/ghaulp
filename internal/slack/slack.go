package slack

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/go-github/v62/github"
)

type payload struct {
	Text string `json:"text"`
}

func PostEvents(webhookURL string, events []*github.AuditEntry, installationIDs map[string]int64, org string) error {
	for i := len(events) - 1; i >= 0; i-- {
		msg := FormatMessage(*events[i], installationIDs, org)
		if err := post(webhookURL, msg); err != nil {
			return fmt.Errorf("posting event to Slack: %w", err)
		}
	}

	return nil
}

func post(webhookURL, message string) error {
	body, err := json.Marshal(payload{Text: message})
	if err != nil {
		return fmt.Errorf("marshalling slack payload: %w", err)
	}

	resp, err := http.Post(webhookURL, "application/json", bytes.NewReader(body)) // #nosec
	if err != nil {
		return fmt.Errorf("posting to slack: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("slack returned non-200 status %d: %s", resp.StatusCode, respBody)
	}
	return nil
}

// FormatMessage builds a Slack message from an audit entry.
// For integration.update events an actionable link to the permissions approval page is appended.
func FormatMessage(e github.AuditEntry, installationIDs map[string]int64, org string) string {
	action := stringVal(e.Action)
	actor := stringVal(e.Actor)

	// GitHub surfaces the GitHub App name in different keys depending on the
	// event; try the most common ones in order.
	appName := additionalString(e, "name")
	if appName == "" {
		appName = additionalString(e, "integration")
	}
	if appName == "" {
		appName = additionalString(e, "app_slug")
	}
	if appName == "" {
		appName = "unknown"
	}

	ts := ""
	if e.Timestamp != nil {
		ts = e.Timestamp.Format(time.RFC3339)
	}

	msg := fmt.Sprintf("*%s* by `%s` — app: `%s` (%s)", action, actor, appName, ts)

	if action == "integration.update" {
		if id, ok := installationIDs[appName]; ok {
			msg += fmt.Sprintf("\n<https://github.com/organizations/%s/settings/installations/%d/permissions/update|Review and approve permissions>", org, id)
		}
	}

	return msg
}

// additionalString extracts a string value from AuditEntry.AdditionalFields or AuditEntry.Data by key.
func additionalString(e github.AuditEntry, key string) string {
	if v, ok := e.AdditionalFields[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if v, ok := e.Data[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func stringVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
