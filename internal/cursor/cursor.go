package cursor

import (
	"context"
	"fmt"
	"io"
	"strings"

	"cloud.google.com/go/storage"
)

const objectName = "cursor.txt"

// Read fetches the last-seen audit log cursor from GCS.
// Returns empty string when the object does not exist yet (first run).
func Read(ctx context.Context, bucket string) (string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return "", fmt.Errorf("creating GCS client: %w", err)
	}
	defer client.Close()

	rc, err := client.Bucket(bucket).Object(objectName).NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return "", nil
		}

		return "", fmt.Errorf("reading cursor from GCS: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return "", fmt.Errorf("reading cursor data: %w", err)
	}

	return strings.TrimSpace(string(data)), nil
}

// Write persists the newest audit log cursor to GCS.
func Write(ctx context.Context, bucket, cursor string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("creating GCS client: %w", err)
	}
	defer client.Close()

	wc := client.Bucket(bucket).Object(objectName).NewWriter(ctx)
	if _, err := wc.Write([]byte(cursor)); err != nil {
		return fmt.Errorf("writing cursor: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("closing cursor writer: %w", err)
	}

	return nil
}
