package filesystem

import "context"

type Filesystem struct {
	path string
}

func NewFilesystem(path string) *Filesystem {
	return &Filesystem{
		path: path,
	}
}

func (f *Filesystem) Path() string {
	return f.path
}

func (f *Filesystem) ListObjects(ctx context.Context, path string) ([]string, error) {
	return nil, nil
}

func (f *Filesystem) RemoveObject(ctx context.Context, path, key string) error {
	return nil
}
