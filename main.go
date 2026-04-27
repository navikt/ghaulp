package main

import (
	"context"
	"log"

	"github.com/navikt/ghaulp/internal/bucket"
	"github.com/navikt/ghaulp/internal/config"
	"github.com/navikt/ghaulp/internal/github"
	"github.com/navikt/ghaulp/internal/slack"
)

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	timestamp, err := bucket.Read(ctx, cfg.GCSBucket)
	if err != nil {
		return err
	}

	log.Printf("Starting run. Stored cursor: %q", timestamp)
	events, newestTimestamp, err := github.FetchEvents(ctx, cfg, timestamp)
	if err != nil {
		return err
	}

	if len(events) == 0 {
		log.Println("No new events. Exiting.")
		return nil
	}

	log.Printf("Fetched %d matching events", len(events))

	if err := slack.PostEvents(cfg.SlackWebhook, events); err != nil {
		return err
	}

	if err := bucket.Write(ctx, cfg.GCSBucket, newestTimestamp.UnixMilli()); err != nil {
		return err
	}

	log.Printf("Cursor updated to %q", newestTimestamp)

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("ghaulp: %v", err)
	}
}
