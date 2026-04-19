package outbound

import (
	"context"
	"io"
)

type StorageService interface {
	Upload(ctx context.Context, key string, content io.Reader) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	GenerateUploadURL(ctx context.Context, key string) (string, error)
}
