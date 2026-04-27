package bucket

import (
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

const objectName = "timestamp.txt"

// Read fetches the last-seen audit log timestamp from GCS.
// Returns empty string when the object does not exist yet (first run).
func Read(ctx context.Context, bucket string) (time.Time, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return time.Time{}, fmt.Errorf("creating GCS client: %w", err)
	}
	defer client.Close()

	rc, err := client.Bucket(bucket).Object(objectName).NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return time.Time{}, nil
		}

		return time.Time{}, fmt.Errorf("reading timestamp from GCS: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return time.Time{}, fmt.Errorf("reading timestamp data: %w", err)
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(timestamp/1000, 0), nil
}

// Write persists the newest audit log timestamp to GCS.
func Write(ctx context.Context, bucket string, timestamp int64) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("creating GCS client: %w", err)
	}
	defer client.Close()

	wc := client.Bucket(bucket).Object(objectName).NewWriter(ctx)
	if _, err := wc.Write([]byte(strconv.FormatInt(timestamp, 10))); err != nil {
		return fmt.Errorf("writing timestamp: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("closing timestamp writer: %w", err)
	}

	return nil
}
