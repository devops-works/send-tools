package backends

import "context"

type Storage interface {
	ListObjects(ctx context.Context, path string) ([]string, error)
	RemoveObject(ctx context.Context, key string) error
	Path() string
}
