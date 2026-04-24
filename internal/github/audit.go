package github

import (
	"context"
	"fmt"

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
func FetchEvents(ctx context.Context, cfg config.Config, after string) ([]github.AuditEntry, string, error) {
	var collected []*github.AuditEntry
	var newestCursor string

	token, err := NewInstallationToken(ctx, cfg.GithubAppID, cfg.GithubClientID, cfg.GithubAppPrivateKey, cfg.GithubOrg)
	if err != nil {
		return nil, "", fmt.Errorf("obtaining GitHub installation token: %w", err)
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	for _, phrase := range []string{
		"action:" + EventInstallation,
		"action:" + EventInstallationRequest,
	} {
		opts := &github.GetAuditLogOptions{
			Phrase: new(phrase),
			Order:  new("desc"),
			ListCursorOptions: github.ListCursorOptions{
				PerPage: 100,
			},
		}

		if after != "" {
			opts.ListCursorOptions.After = after
		}

		for {
			entries, resp, err := client.Organizations.GetAuditLog(ctx, cfg.GithubOrg, opts)
			if err != nil {
				return nil, "", fmt.Errorf("fetching audit log (phrase=%q): %w", phrase, err)
			}

			// Capture the cursor from the very first page of the first query
			// as the new high-water mark.
			if newestCursor == "" && resp.Before != "" {
				newestCursor = resp.Before
			}

			collected = append(collected, entries...)

			if resp.After == "" {
				break
			}
			opts.ListCursorOptions.After = resp.After
		}
	}

	var filtered []github.AuditEntry
	for _, e := range collected {
		if e.Action != nil && (*e.Action == EventInstallation || *e.Action == EventInstallationRequest) {
			filtered = append(filtered, *e)
		}
	}

	return filtered, newestCursor, nil
}
