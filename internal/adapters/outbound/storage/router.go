package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/bbridges_11/document-registry/internal/domain/shared"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// Router implements StorageService by delegating to backend-specific adapters
type Router struct {
	backends map[shared.StorageBackend]outbound.StorageBackend
}

// NewRouter creates a storage router with registered backends
func NewRouter(backends ...outbound.StorageBackend) *Router {
	backendMap := make(map[shared.StorageBackend]outbound.StorageBackend)
	for _, backend := range backends {
		backendMap[backend.Type()] = backend
	}
	return &Router{backends: backendMap}
}

func (r *Router) Upload(ctx context.Context, req outbound.StorageRequest) (ref *shared.ContentReference, err error) {
	defer err2.Handle(&err)

	backend := try.To1(r.getBackend(req.Backend))
	try.To(backend.Upload(ctx, req.Key, req.Content))

	return shared.NewContentReference(req.Backend, req.Key)
}

func (r *Router) Download(ctx context.Context, ref *shared.ContentReference) (_ io.ReadCloser, err error) {
	defer err2.Handle(&err)

	backend := try.To1(r.getBackend(ref.Backend()))
	return backend.Download(ctx, ref.Location())
}

func (r *Router) Delete(ctx context.Context, ref *shared.ContentReference) (err error) {
	defer err2.Handle(&err)

	backend := try.To1(r.getBackend(ref.Backend()))
	return backend.Delete(ctx, ref.Location())
}

func (r *Router) Exists(ctx context.Context, ref *shared.ContentReference) (exists bool, err error) {
	defer err2.Handle(&err)

	backend := try.To1(r.getBackend(ref.Backend()))
	return backend.Exists(ctx, ref.Location())
}

func (r *Router) GenerateUploadURL(ctx context.Context, req outbound.StorageRequest) (url string, err error) {
	defer err2.Handle(&err)

	backend := try.To1(r.getBackend(req.Backend))
	return backend.GenerateUploadURL(ctx, req.Key)
}

func (r *Router) getBackend(backendType shared.StorageBackend) (outbound.StorageBackend, error) {
	backend, ok := r.backends[backendType]
	if !ok {
		return nil, errors.New(
			errors.CodeNotFound,
			fmt.Sprintf("storage backend not configured: %s", backendType),
		)
	}
	return backend, nil
}
