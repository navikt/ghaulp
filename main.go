package main

import (
	"context"
	"log"

	"github.com/navikt/ghaulp/internal/config"
	"github.com/navikt/ghaulp/internal/cursor"
	"github.com/navikt/ghaulp/internal/github"
	"github.com/navikt/ghaulp/internal/slack"
)

func run() error {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	cur, err := cursor.Read(ctx, cfg.GCSBucket)
	if err != nil {
		return err
	}

	log.Printf("Starting run. Stored cursor: %q", cur)
	events, newestCursor, err := github.FetchEvents(ctx, cfg, cur)
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

	if newestCursor != "" && newestCursor != cur {
		if err := cursor.Write(ctx, cfg.GCSBucket, newestCursor); err != nil {
			return err
		}

		log.Printf("Cursor updated to %q", newestCursor)
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatalf("ghaulp: %v", err)
	}
}
