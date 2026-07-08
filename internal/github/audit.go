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
	EventIntegrationUpdate   = "integration.update"
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

	searchPhrase := "action:integration_installation action:integration_installation_request action:integration.update -action:integration_installation.repositories_added -action:integration_installation.repositories_removed"
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

// ResolveInstallationIDs fetches all GitHub App installations for the org and returns
// a map of app slug → installation ID. This is used to build approval links for
// integration.update events, which do not carry an installation ID themselves.
func ResolveInstallationIDs(ctx context.Context, cfg config.Config) (map[string]int64, error) {
	token, err := GenerateAccessToken(ctx, cfg.GithubClientID, cfg.GithubInstallationID, cfg.GithubAppPrivateKey, cfg.GithubOrg)
	if err != nil {
		return nil, fmt.Errorf("obtaining GitHub installation token for installation lookup: %w", err)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	result, _, err := client.Organizations.ListInstallations(ctx, cfg.GithubOrg, nil)
	if err != nil {
		return nil, fmt.Errorf("listing org installations: %w", err)
	}

	ids := make(map[string]int64, len(result.Installations))
	for _, inst := range result.Installations {
		if inst.AppSlug != nil && inst.ID != nil {
			ids[*inst.AppSlug] = *inst.ID
		}
	}
	return ids, nil
}
