package postgres

import (
	"context"
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/publication"
	dbPostgres "github.com/bbridges_11/document-registry/internal/platform/database/postgres"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// PublicationRepository implements outbound.PublicationRepository
type PublicationRepository struct {
	runner *dbPostgres.Runner
}

// NewPublicationRepository creates a new publication repository
func NewPublicationRepository(runner *dbPostgres.Runner) *PublicationRepository {
	return &PublicationRepository{runner: runner}
}

// Save persists a publication
func (r *PublicationRepository) Save(ctx context.Context, pub *publication.Publication) (err error) {
	defer err2.Handle(&err)

	query := `
		INSERT INTO publications (
			id, version_id, document_id, published_to, published_by, published_at, status, error_message
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	querier := r.runner.GetQuerier(ctx)
	try.To1(querier.Exec(ctx, query,
		pub.ID(),
		pub.VersionID(),
		pub.DocumentID(),
		pub.PublishedTo(),
		pub.PublishedBy(),
		pub.PublishedAt(),
		string(pub.Status()),
		pub.ErrorMessage(),
	))

	return nil
}

// GetByID retrieves a publication by ID
func (r *PublicationRepository) GetByID(ctx context.Context, id uuid.UUID) (pub *publication.Publication, err error) {
	defer err2.Handle(&err)

	query := `
		SELECT id, version_id, document_id, published_to, published_by, published_at, status, error_message
		FROM publications
		WHERE id = $1
	`

	querier := r.runner.GetQuerier(ctx)
	row := querier.QueryRow(ctx, query, id)

	var (
		pubID        uuid.UUID
		versionID    uuid.UUID
		documentID   uuid.UUID
		publishedTo  string
		publishedBy  string
		publishedAt  time.Time
		status       string
		errorMessage string
	)

	err = row.Scan(&pubID, &versionID, &documentID, &publishedTo, &publishedBy, &publishedAt, &status, &errorMessage)
	if err == pgx.ErrNoRows {
		return nil, errors.New(errors.CodeNotFound, "publication not found")
	}
	try.To(err)

	return publication.RehydratePublication(
		pubID,
		versionID,
		documentID,
		publishedTo,
		publishedBy,
		publishedAt,
		publication.PublicationStatus(status),
		errorMessage,
	), nil
}

// ListByVersionID retrieves all publications for a version
func (r *PublicationRepository) ListByVersionID(ctx context.Context, versionID uuid.UUID) (pubs []*publication.Publication, err error) {
	defer err2.Handle(&err)

	query := `
		SELECT id, version_id, document_id, published_to, published_by, published_at, status, error_message
		FROM publications
		WHERE version_id = $1
		ORDER BY published_at DESC
	`

	querier := r.runner.GetQuerier(ctx)
	rows := try.To1(querier.Query(ctx, query, versionID))
	defer rows.Close()

	publications := []*publication.Publication{}
	for rows.Next() {
		var (
			pubID        uuid.UUID
			vID          uuid.UUID
			documentID   uuid.UUID
			publishedTo  string
			publishedBy  string
			publishedAt  time.Time
			status       string
			errorMessage string
		)

		try.To(rows.Scan(&pubID, &vID, &documentID, &publishedTo, &publishedBy, &publishedAt, &status, &errorMessage))

		pub := publication.RehydratePublication(
			pubID,
			vID,
			documentID,
			publishedTo,
			publishedBy,
			publishedAt,
			publication.PublicationStatus(status),
			errorMessage,
		)
		publications = append(publications, pub)
	}

	try.To(rows.Err())
	return publications, nil
}

// ListAll retrieves all publications with pagination
func (r *PublicationRepository) ListAll(ctx context.Context, limit int, offset int) (pubs []*publication.Publication, err error) {
	defer err2.Handle(&err)

	query := `
		SELECT id, version_id, document_id, published_to, published_by, published_at, status, error_message
		FROM publications
		ORDER BY published_at DESC
		LIMIT $1 OFFSET $2
	`

	querier := r.runner.GetQuerier(ctx)
	rows := try.To1(querier.Query(ctx, query, limit, offset))
	defer rows.Close()

	publications := []*publication.Publication{}
	for rows.Next() {
		var (
			pubID        uuid.UUID
			versionID    uuid.UUID
			documentID   uuid.UUID
			publishedTo  string
			publishedBy  string
			publishedAt  time.Time
			status       string
			errorMessage string
		)

		try.To(rows.Scan(&pubID, &versionID, &documentID, &publishedTo, &publishedBy, &publishedAt, &status, &errorMessage))

		pub := publication.RehydratePublication(
			pubID,
			versionID,
			documentID,
			publishedTo,
			publishedBy,
			publishedAt,
			publication.PublicationStatus(status),
			errorMessage,
		)
		publications = append(publications, pub)
	}

	try.To(rows.Err())
	return publications, nil
}

// Count returns the total number of publications
func (r *PublicationRepository) Count(ctx context.Context) (count int, err error) {
	defer err2.Handle(&err)

	query := `SELECT COUNT(*) FROM publications`

	querier := r.runner.GetQuerier(ctx)
	row := querier.QueryRow(ctx, query)

	try.To(row.Scan(&count))
	return count, nil
}

// ListByDocumentID retrieves all publications for a document
func (r *PublicationRepository) ListByDocumentID(ctx context.Context, documentID uuid.UUID) (pubs []*publication.Publication, err error) {
	defer err2.Handle(&err)

	query := `
		SELECT id, version_id, document_id, published_to, published_by, published_at, status, error_message
		FROM publications
		WHERE document_id = $1
		ORDER BY published_at DESC
	`

	querier := r.runner.GetQuerier(ctx)
	rows := try.To1(querier.Query(ctx, query, documentID))
	defer rows.Close()

	publications := []*publication.Publication{}
	for rows.Next() {
		var (
			pubID        uuid.UUID
			versionID    uuid.UUID
			docID        uuid.UUID
			publishedTo  string
			publishedBy  string
			publishedAt  time.Time
			status       string
			errorMessage string
		)

		try.To(rows.Scan(&pubID, &versionID, &docID, &publishedTo, &publishedBy, &publishedAt, &status, &errorMessage))

		pub := publication.RehydratePublication(
			pubID,
			versionID,
			docID,
			publishedTo,
			publishedBy,
			publishedAt,
			publication.PublicationStatus(status),
			errorMessage,
		)
		publications = append(publications, pub)
	}

	try.To(rows.Err())
	return publications, nil
}
