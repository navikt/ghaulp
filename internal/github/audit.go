package github

import (
	"context"
	"fmt"
	"time"

	"github.com/google/go-github/v62/github"
	"github.com/navikt/ghaulp/internal/config"
	"golang.org/x/oauth2"
)

const (
	EventInstallation        = "integration_installation"
	EventInstallationRequest = "integration_installation_request"
)

// FetchEvents retrieves audit log entries for the two integration event types,
// stopping pagination once the stored cursor is encountered or pages run out.
// Returned events are ordered newest-first (as returned by the API).
func FetchEvents(ctx context.Context, cfg config.Config, timestamp time.Time) ([]*github.AuditEntry, time.Time, error) {
	token, err := GenerateAccessToken(ctx, cfg.GithubClientID, cfg.GithubInstallationID, cfg.GithubAppPrivateKey, cfg.GithubOrg)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("obtaining GitHub installation token: %w", err)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	searchPhrase := "action:integration_installation action:integration_installation_request -action:integration_installation.repositories_added  -action:integration_installation.repositories_removed"
	opts := &github.GetAuditLogOptions{
		Phrase: new(searchPhrase),
		Order:  new("desc"),
		ListCursorOptions: github.ListCursorOptions{
			PerPage: 100,
		},
	}

	entries, _, err := client.Organizations.GetAuditLog(ctx, cfg.GithubOrg, opts)
	if err != nil {
		return nil, time.Time{}, fmt.Errorf("fetching audit log (phrase=%q): %w", searchPhrase, err)
	}

	var collected []*github.AuditEntry
	for _, entry := range entries {
		if entry.Timestamp.After(timestamp) {
			collected = append(collected, entry)
		}
	}

	var newestTimestamp time.Time
	if len(collected) > 0 {
		newestTimestamp = collected[0].GetTimestamp().Time
	}

	return collected, newestTimestamp, nil
}
