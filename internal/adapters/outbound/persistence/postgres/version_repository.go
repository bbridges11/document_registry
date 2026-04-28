package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/shared"
	"github.com/bbridges_11/document-registry/internal/domain/version"
	"github.com/bbridges_11/document-registry/internal/domain/workflow"
	dbPostgres "github.com/bbridges_11/document-registry/internal/platform/database/postgres"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type VersionRepository struct {
	runner *dbPostgres.Runner
}

func NewVersionRepository(runner *dbPostgres.Runner) *VersionRepository {
	return &VersionRepository{runner: runner}
}

func (r *VersionRepository) Save(ctx context.Context, ver *version.Version) (err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	metadataJSON := try.To1(json.Marshal(ver.Metadata()))
	ref := ver.ContentRef()

	query := `
		INSERT INTO versions (id, document_id, version, status, content_backend, content_location, content_hash, metadata, created_by, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			content_backend = EXCLUDED.content_backend,
			content_location = EXCLUDED.content_location,
			content_hash = EXCLUDED.content_hash,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
	`

	try.To1(q.Exec(ctx, query,
		ver.ID(),
		ver.DocumentID(),
		ver.Version().String(),
		ver.Status(),
		string(ref.Backend()),
		ref.Location(),
		ver.ContentHash(),
		metadataJSON,
		ver.CreatedBy(),
		ver.CreatedAt(),
		ver.UpdatedAt(),
	))

	return nil
}

func (r *VersionRepository) GetByID(ctx context.Context, id uuid.UUID) (ver *version.Version, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT id, document_id, version, status, content_backend, content_location, content_hash, metadata, created_by, created_at, updated_at
		FROM versions
		WHERE id = $1
	`

	var (
		verID           uuid.UUID
		documentID      uuid.UUID
		versionStr      string
		status          workflow.Status
		contentBackend  string
		contentLocation string
		contentHash     string
		metadataJSON    []byte
		createdBy       string
		createdAt       time.Time
		updatedAt       time.Time
	)

	err = q.QueryRow(ctx, query, id).Scan(
		&verID, &documentID, &versionStr, &status, &contentBackend, &contentLocation, &contentHash, &metadataJSON, &createdBy, &createdAt, &updatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	try.To(err)

	var metadata map[string]any
	try.To(json.Unmarshal(metadataJSON, &metadata))

	contentRef := try.To1(shared.NewContentReference(shared.StorageBackend(contentBackend), contentLocation))

	return version.RehydrateVersion(verID, documentID, versionStr, status, contentRef, contentHash, createdBy, metadata, createdAt, updatedAt)
}

func (r *VersionRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (ver *version.Version, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT id, document_id, version, status, content_backend, content_location, content_hash, metadata, created_by, created_at, updated_at
		FROM versions
		WHERE id = $1
		FOR UPDATE
	`

	var (
		verID           uuid.UUID
		documentID      uuid.UUID
		versionStr      string
		status          workflow.Status
		contentBackend  string
		contentLocation string
		contentHash     string
		metadataJSON    []byte
		createdBy       string
		createdAt       time.Time
		updatedAt       time.Time
	)

	err = q.QueryRow(ctx, query, id).Scan(
		&verID, &documentID, &versionStr, &status, &contentBackend, &contentLocation, &contentHash, &metadataJSON, &createdBy, &createdAt, &updatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	try.To(err)

	var metadata map[string]any
	try.To(json.Unmarshal(metadataJSON, &metadata))

	contentRef := try.To1(shared.NewContentReference(shared.StorageBackend(contentBackend), contentLocation))

	return version.RehydrateVersion(verID, documentID, versionStr, status, contentRef, contentHash, createdBy, metadata, createdAt, updatedAt)
}

func (r *VersionRepository) GetByDocumentIDAndVersion(ctx context.Context, documentID uuid.UUID, versionStr string) (ver *version.Version, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT id, document_id, version, status, content_backend, content_location, content_hash, metadata, created_by, created_at, updated_at
		FROM versions
		WHERE document_id = $1 AND version = $2
	`

	var (
		verID           uuid.UUID
		docID           uuid.UUID
		version_        string
		status          workflow.Status
		contentBackend  string
		contentLocation string
		contentHash     string
		metadataJSON    []byte
		createdBy       string
		createdAt       time.Time
		updatedAt       time.Time
	)

	err = q.QueryRow(ctx, query, documentID, versionStr).Scan(
		&verID, &docID, &version_, &status, &contentBackend, &contentLocation, &contentHash, &metadataJSON, &createdBy, &createdAt, &updatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	try.To(err)

	var metadata map[string]any
	try.To(json.Unmarshal(metadataJSON, &metadata))

	contentRef := try.To1(shared.NewContentReference(shared.StorageBackend(contentBackend), contentLocation))

	return version.RehydrateVersion(verID, docID, version_, status, contentRef, contentHash, createdBy, metadata, createdAt, updatedAt)
}

func (r *VersionRepository) ListByDocumentID(ctx context.Context, documentID uuid.UUID) (versions []*version.Version, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT id, document_id, version, status, content_backend, content_location, content_hash, metadata, created_by, created_at, updated_at
		FROM versions
		WHERE document_id = $1
		ORDER BY created_at DESC
	`

	rows := try.To1(q.Query(ctx, query, documentID))
	defer rows.Close()

	versions = make([]*version.Version, 0)

	for rows.Next() {
		var (
			verID           uuid.UUID
			docID           uuid.UUID
			versionStr      string
			status          workflow.Status
			contentBackend  string
			contentLocation string
			contentHash     string
			metadataJSON    []byte
			createdBy       string
			createdAt       time.Time
			updatedAt       time.Time
		)

		try.To(rows.Scan(&verID, &docID, &versionStr, &status, &contentBackend, &contentLocation, &contentHash, &metadataJSON, &createdBy, &createdAt, &updatedAt))

		var metadata map[string]any
		try.To(json.Unmarshal(metadataJSON, &metadata))

		contentRef := try.To1(shared.NewContentReference(shared.StorageBackend(contentBackend), contentLocation))

		ver := try.To1(version.RehydrateVersion(verID, docID, versionStr, status, contentRef, contentHash, createdBy, metadata, createdAt, updatedAt))
		versions = append(versions, ver)
	}

	try.To(rows.Err())

	return versions, nil
}

func (r *VersionRepository) GetLatestByDocumentID(ctx context.Context, documentID uuid.UUID) (ver *version.Version, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT id, document_id, version, status, content_backend, content_location, content_hash, metadata, created_by, created_at, updated_at
		FROM versions
		WHERE document_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	var (
		verID           uuid.UUID
		docID           uuid.UUID
		versionStr      string
		status          workflow.Status
		contentBackend  string
		contentLocation string
		contentHash     string
		metadataJSON    []byte
		createdBy       string
		createdAt       time.Time
		updatedAt       time.Time
	)

	err = q.QueryRow(ctx, query, documentID).Scan(
		&verID, &docID, &versionStr, &status, &contentBackend, &contentLocation, &contentHash, &metadataJSON, &createdBy, &createdAt, &updatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	try.To(err)

	var metadata map[string]any
	try.To(json.Unmarshal(metadataJSON, &metadata))

	contentRef := try.To1(shared.NewContentReference(shared.StorageBackend(contentBackend), contentLocation))

	return version.RehydrateVersion(verID, docID, versionStr, status, contentRef, contentHash, createdBy, metadata, createdAt, updatedAt)
}

func (r *VersionRepository) GetPublishedByDocumentID(ctx context.Context, documentID uuid.UUID) (ver *version.Version, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT id, document_id, version, status, content_backend, content_location, content_hash, metadata, created_by, created_at, updated_at
		FROM versions
		WHERE document_id = $1 AND status = $2
		LIMIT 1
	`

	var (
		verID           uuid.UUID
		docID           uuid.UUID
		versionStr      string
		status          workflow.Status
		contentBackend  string
		contentLocation string
		contentHash     string
		metadataJSON    []byte
		createdBy       string
		createdAt       time.Time
		updatedAt       time.Time
	)

	err = q.QueryRow(ctx, query, documentID, workflow.StatusPublished).Scan(
		&verID, &docID, &versionStr, &status, &contentBackend, &contentLocation, &contentHash, &metadataJSON, &createdBy, &createdAt, &updatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	try.To(err)

	var metadata map[string]any
	try.To(json.Unmarshal(metadataJSON, &metadata))

	contentRef := try.To1(shared.NewContentReference(shared.StorageBackend(contentBackend), contentLocation))

	return version.RehydrateVersion(verID, docID, versionStr, status, contentRef, contentHash, createdBy, metadata, createdAt, updatedAt)
}

func (r *VersionRepository) HasPublishedVersion(ctx context.Context, documentID uuid.UUID) (hasPublished bool, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `SELECT EXISTS(SELECT 1 FROM versions WHERE document_id = $1 AND status = $2)`
	try.To(q.QueryRow(ctx, query, documentID, workflow.StatusPublished).Scan(&hasPublished))

	return hasPublished, nil
}

func (r *VersionRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status workflow.Status) (err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `UPDATE versions SET status = $1, updated_at = $2 WHERE id = $3`
	try.To1(q.Exec(ctx, query, status, time.Now().UTC(), id))

	return nil
}

func (r *VersionRepository) ExistsByDocumentIDAndVersion(ctx context.Context, documentID uuid.UUID, versionStr string) (exists bool, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `SELECT EXISTS(SELECT 1 FROM versions WHERE document_id = $1 AND version = $2)`
	try.To(q.QueryRow(ctx, query, documentID, versionStr).Scan(&exists))

	return exists, nil
}
