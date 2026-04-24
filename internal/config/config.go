package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	GithubAppID         string
	GithubAppPrivateKey string
	GithubClientID      string
	GithubOrg           string
	SlackWebhook        string
	GCSBucket           string
}

func Load() (Config, error) {
	cfg := Config{
		GithubAppID:         os.Getenv("GITHUB_APP_ID"),
		GithubAppPrivateKey: os.Getenv("GITHUB_APP_PRIVATE_KEY"),
		GithubClientID:      os.Getenv("GITHUB_CLIENT_ID"),
		GithubOrg:           os.Getenv("GITHUB_ORG"),
		SlackWebhook:        os.Getenv("SLACK_WEBHOOK_URL"),
		GCSBucket:           os.Getenv("GCS_BUCKET"),
	}
	var missing []string
	if cfg.GithubAppID == "" {
		missing = append(missing, "GITHUB_APP_ID")
	}
	if cfg.GithubAppPrivateKey == "" {
		missing = append(missing, "GITHUB_APP_PRIVATE_KEY")
	}
	if cfg.GithubClientID == "" {
		missing = append(missing, "GITHUB_CLIENT_ID")
	}
	if cfg.GithubOrg == "" {
		missing = append(missing, "GITHUB_ORG")
	}
	if cfg.SlackWebhook == "" {
		missing = append(missing, "SLACK_WEBHOOK_URL")
	}
	if cfg.GCSBucket == "" {
		missing = append(missing, "GCS_BUCKET")
	}
	if len(missing) > 0 {
		return cfg, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return cfg, nil
}
