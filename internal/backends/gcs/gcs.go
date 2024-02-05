package gcs

import (
	"context"
	"fmt"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type GCS struct {
	bucket string
}

func NewGCS(bucket string) *GCS {
	return &GCS{
		bucket: bucket,
	}
}

func (g *GCS) Path() string {
	return g.bucket
}

func (g *GCS) ListObjects(ctx context.Context, path string) ([]string, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	var files []string
	it := client.Bucket(g.bucket).Objects(ctx, nil)
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		files = append(files, attrs.Name)
	}
	return files, nil

}

func (g *GCS) RemoveObject(ctx context.Context, key string) error {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return fmt.Errorf("unable to create GCS client: %w", err)
	}
	defer client.Close()

	// remove file from GCS
	if err := client.Bucket(g.bucket).Object(key).Delete(ctx); err != nil {
		return fmt.Errorf("unable to remove key %s from GCS: %w", key, err)

	}
	return nil
}
