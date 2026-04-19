package postgres

import (
	"context"
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/deprecation"
	dbPostgres "github.com/bbridges_11/document-registry/internal/platform/database/postgres"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// DeprecationRepository implements the deprecation repository using Postgres
type DeprecationRepository struct {
	runner *dbPostgres.Runner
}

// NewDeprecationRepository creates a new Postgres deprecation repository
func NewDeprecationRepository(runner *dbPostgres.Runner) *DeprecationRepository {
	return &DeprecationRepository{runner: runner}
}

// Save persists a deprecation entity
func (r *DeprecationRepository) Save(ctx context.Context, dep *deprecation.Deprecation) (err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		INSERT INTO deprecations (
			id, version_id, document_id, requested_by, requested_at, reason, deprecation_note,
			auto_deprecated, status, deprecated_by, deprecated_at, superseded_by, previous_status,
			created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			deprecated_by = EXCLUDED.deprecated_by,
			deprecated_at = EXCLUDED.deprecated_at,
			updated_at = EXCLUDED.updated_at
	`

	try.To1(q.Exec(ctx, query,
		dep.ID(),
		dep.VersionID(),
		dep.DocumentID(),
		dep.RequestedBy(),
		dep.RequestedAt(),
		dep.Reason(),
		nullString(dep.DeprecationNote()),
		dep.AutoDeprecated(),
		dep.Status(),
		nullString(dep.DeprecatedBy()),
		dep.DeprecatedAt(),
		nullUUID(dep.SupersededBy()),
		dep.PreviousStatus(),
		dep.CreatedAt(),
		dep.UpdatedAt(),
	))

	return nil
}

// GetByID retrieves a deprecation by its ID
func (r *DeprecationRepository) GetByID(ctx context.Context, id uuid.UUID) (dep *deprecation.Deprecation, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT 
			id, version_id, document_id, requested_by, requested_at, reason, deprecation_note,
			auto_deprecated, status, deprecated_by, deprecated_at, superseded_by, previous_status,
			created_at, updated_at
		FROM deprecations
		WHERE id = $1
	`

	row := q.QueryRow(ctx, query, id)
	dep = try.To1(scanDeprecation(row))

	return dep, nil
}

// GetByVersionID retrieves the most recent deprecation for a version
func (r *DeprecationRepository) GetByVersionID(ctx context.Context, versionID uuid.UUID) (dep *deprecation.Deprecation, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT 
			id, version_id, document_id, requested_by, requested_at, reason, deprecation_note,
			auto_deprecated, status, deprecated_by, deprecated_at, superseded_by, previous_status,
			created_at, updated_at
		FROM deprecations
		WHERE version_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := q.QueryRow(ctx, query, versionID)
	dep = try.To1(scanDeprecation(row))

	return dep, nil
}

// GetPendingByVersionID retrieves a pending deprecation for a version (if exists)
func (r *DeprecationRepository) GetPendingByVersionID(ctx context.Context, versionID uuid.UUID) (*deprecation.Deprecation, error) {
	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT 
			id, version_id, document_id, requested_by, requested_at, reason, deprecation_note,
			auto_deprecated, status, deprecated_by, deprecated_at, superseded_by, previous_status,
			created_at, updated_at
		FROM deprecations
		WHERE version_id = $1 AND status = 'PENDING'
		LIMIT 1
	`

	row := q.QueryRow(ctx, query, versionID)
	dep, err := scanDeprecation(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // No pending deprecation found (not an error)
		}
		return nil, err
	}

	return dep, nil
}

// ListByDocumentID retrieves all deprecations for a document
func (r *DeprecationRepository) ListByDocumentID(ctx context.Context, documentID uuid.UUID) (deps []*deprecation.Deprecation, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT 
			id, version_id, document_id, requested_by, requested_at, reason, deprecation_note,
			auto_deprecated, status, deprecated_by, deprecated_at, superseded_by, previous_status,
			created_at, updated_at
		FROM deprecations
		WHERE document_id = $1
		ORDER BY created_at DESC
	`

	rows := try.To1(q.Query(ctx, query, documentID))
	defer rows.Close()

	deps = []*deprecation.Deprecation{}
	for rows.Next() {
		dep := try.To1(scanDeprecation(rows))
		deps = append(deps, dep)
	}

	return deps, nil
}

// scanDeprecation scans a row into a deprecation entity
func scanDeprecation(row pgx.Row) (*deprecation.Deprecation, error) {
	var (
		id              uuid.UUID
		versionID       uuid.UUID
		documentID      uuid.UUID
		requestedBy     string
		requestedAt     time.Time
		reason          string
		deprecationNote *string
		autoDeprecated  bool
		status          string
		deprecatedBy    *string
		deprecatedAt    *time.Time
		supersededBy    *uuid.UUID
		previousStatus  string
		createdAt       time.Time
		updatedAt       time.Time
	)

	err := row.Scan(
		&id, &versionID, &documentID, &requestedBy, &requestedAt, &reason, &deprecationNote,
		&autoDeprecated, &status, &deprecatedBy, &deprecatedAt, &supersededBy, &previousStatus,
		&createdAt, &updatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New(errors.CodeNotFound, "deprecation not found")
		}
		return nil, err
	}

	dep := deprecation.RehydrateDeprecation(
		id,
		versionID,
		documentID,
		requestedBy,
		requestedAt,
		reason,
		stringValue(deprecationNote),
		autoDeprecated,
		deprecation.DeprecationStatus(status),
		stringValue(deprecatedBy),
		deprecatedAt,
		supersededBy,
		previousStatus,
		createdAt,
		updatedAt,
	)

	return dep, nil
}

// Helper functions for nullable types

func nullString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func nullUUID(u *uuid.UUID) *uuid.UUID {
	if u == nil {
		return nil
	}
	return u
}
