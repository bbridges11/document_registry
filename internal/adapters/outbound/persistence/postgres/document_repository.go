package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/document"
	dbPostgres "github.com/bbridges_11/document-registry/internal/platform/database/postgres"
	"github.com/bbridges_11/document-registry/internal/ports/outbound"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type DocumentRepository struct {
	runner *dbPostgres.Runner
}

func NewDocumentRepository(runner *dbPostgres.Runner) *DocumentRepository {
	return &DocumentRepository{runner: runner}
}

func (r *DocumentRepository) Save(ctx context.Context, doc *document.Document) (err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		INSERT INTO documents (id, name, description, tags, document_type, created_at, updated_at, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			description = EXCLUDED.description,
			tags = EXCLUDED.tags,
			updated_at = EXCLUDED.updated_at
	`

	try.To1(q.Exec(ctx, query,
		doc.ID(),
		doc.Name(),
		doc.Description(),
		doc.Tags(),
		doc.DocumentType(),
		doc.CreatedAt(),
		doc.UpdatedAt(),
		doc.CreatedBy(),
	))

	return nil
}

func (r *DocumentRepository) GetByID(ctx context.Context, id uuid.UUID) (doc *document.Document, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT id, name, description, tags, document_type, created_at, updated_at, created_by
		FROM documents
		WHERE id = $1
	`

	var (
		docID        uuid.UUID
		name         string
		description  string
		tags         []string
		documentType string
		createdAt    time.Time
		updatedAt    time.Time
		createdBy    string
	)

	err = q.QueryRow(ctx, query, id).Scan(
		&docID, &name, &description, &tags, &documentType, &createdAt, &updatedAt, &createdBy,
	)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	try.To(err)

	return document.RehydrateDocument(docID, name, description, documentType, createdBy, tags, createdAt, updatedAt), nil
}

func (r *DocumentRepository) List(ctx context.Context, limit, offset int) (docs []*document.Document, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT id, name, description, tags, document_type, created_at, updated_at, created_by
		FROM documents
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows := try.To1(q.Query(ctx, query, limit, offset))
	defer rows.Close()

	docs = make([]*document.Document, 0)

	for rows.Next() {
		var (
			docID        uuid.UUID
			name         string
			description  string
			tags         []string
			documentType string
			createdAt    time.Time
			updatedAt    time.Time
			createdBy    string
		)

		try.To(rows.Scan(&docID, &name, &description, &tags, &documentType, &createdAt, &updatedAt, &createdBy))
		docs = append(docs, document.RehydrateDocument(docID, name, description, documentType, createdBy, tags, createdAt, updatedAt))
	}

	try.To(rows.Err())

	return docs, nil
}

func (r *DocumentRepository) ExistsByID(ctx context.Context, id uuid.UUID) (exists bool, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `SELECT EXISTS(SELECT 1 FROM documents WHERE id = $1)`
	try.To(q.QueryRow(ctx, query, id).Scan(&exists))

	return exists, nil
}

func (r *DocumentRepository) Search(ctx context.Context, filters outbound.DocumentSearchFilters) (results []outbound.DocumentSearchResult, total int, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	// Build base query with latest version info
	query := `
		SELECT 
			d.id,
			d.name,
			d.description,
			d.document_type,
			d.tags,
			d.created_by,
			d.created_at,
			d.updated_at,
			COALESCE(v.version, '') as latest_version,
			COALESCE(v.status, '') as latest_version_status
		FROM documents d
		LEFT JOIN LATERAL (
			SELECT version, status
			FROM versions
			WHERE document_id = d.id
			ORDER BY created_at DESC
			LIMIT 1
		) v ON true
		WHERE 1=1
	`

	args := []interface{}{}
	argIdx := 1

	// Text search (name OR description)
	if filters.Query != "" {
		query += ` AND (d.name ILIKE $` + fmt.Sprintf("%d", argIdx) + ` OR d.description ILIKE $` + fmt.Sprintf("%d", argIdx) + `)`
		args = append(args, "%"+filters.Query+"%")
		argIdx++
	}

	// Name filter
	if filters.Name != "" {
		query += ` AND d.name ILIKE $` + fmt.Sprintf("%d", argIdx)
		args = append(args, "%"+filters.Name+"%")
		argIdx++
	}

	// Description filter
	if filters.Description != "" {
		query += ` AND d.description ILIKE $` + fmt.Sprintf("%d", argIdx)
		args = append(args, "%"+filters.Description+"%")
		argIdx++
	}

	// Document type filter
	if filters.DocumentType != "" {
		query += ` AND d.document_type = $` + fmt.Sprintf("%d", argIdx)
		args = append(args, filters.DocumentType)
		argIdx++
	}

	// Created by filter
	if filters.CreatedBy != "" {
		query += ` AND d.created_by = $` + fmt.Sprintf("%d", argIdx)
		args = append(args, filters.CreatedBy)
		argIdx++
	}

	// Tags filter (array overlap)
	if len(filters.Tags) > 0 {
		query += ` AND d.tags && $` + fmt.Sprintf("%d", argIdx)
		args = append(args, filters.Tags)
		argIdx++
	}

	// Date range filters
	if filters.CreatedAfter != nil {
		query += ` AND d.created_at >= $` + fmt.Sprintf("%d", argIdx)
		args = append(args, *filters.CreatedAfter)
		argIdx++
	}

	if filters.CreatedBefore != nil {
		query += ` AND d.created_at <= $` + fmt.Sprintf("%d", argIdx)
		args = append(args, *filters.CreatedBefore)
		argIdx++
	}

	// Document IDs filter (for authorization scoping)
	if len(filters.DocumentIDs) > 0 {
		query += ` AND d.id = ANY($` + fmt.Sprintf("%d", argIdx) + `)`
		args = append(args, filters.DocumentIDs)
		argIdx++
	}

	// Special filter: My documents
	if filters.MyDocuments && filters.UserID != "" {
		query += ` AND d.created_by = $` + fmt.Sprintf("%d", argIdx)
		args = append(args, filters.UserID)
		argIdx++
	}

	// Special filter: My stakeholder documents
	if filters.MyStakeholderDocs && filters.UserID != "" {
		query += ` AND EXISTS (
			SELECT 1 FROM stakeholders s
			WHERE s.document_id = d.id
			AND s.user_id = $` + fmt.Sprintf("%d", argIdx) + `
		)`
		args = append(args, filters.UserID)
		argIdx++
	}

	// Special filter: Pending my review
	if filters.PendingMyReview && filters.UserID != "" {
		query += ` AND EXISTS (
			SELECT 1 FROM versions v
			WHERE v.document_id = d.id
			AND v.status = 'UNDER_REVIEW'
		)`
		// Note: Actual permission check would require OpenFGA integration
	}

	// Special filter: Pending my approval
	if filters.PendingMyApproval && filters.UserID != "" {
		query += ` AND EXISTS (
			SELECT 1 FROM versions v
			WHERE v.document_id = d.id
			AND v.status = 'UNDER_REVIEW'
			AND NOT EXISTS (
				SELECT 1 FROM approvals a
				WHERE a.version_id = v.id
				AND a.user_id = $` + fmt.Sprintf("%d", argIdx) + `
			)
		)`
		args = append(args, filters.UserID)
		argIdx++
	}

	// Special filter: Recently published
	if filters.RecentlyPublished {
		query += ` AND EXISTS (
			SELECT 1 FROM versions v
			WHERE v.document_id = d.id
			AND v.status = 'PUBLISHED'
			AND v.updated_at > NOW() - INTERVAL '7 days'
		)`
	}

	// Count total results (before pagination)
	countQuery := `SELECT COUNT(*) FROM (` + query + `) as count_query`
	try.To(q.QueryRow(ctx, countQuery, args...).Scan(&total))

	// Add sorting
	sortField := "d.created_at"
	sortOrder := "DESC"

	if filters.SortField != "" {
		switch filters.SortField {
		case "name":
			sortField = "d.name"
		case "updated_at":
			sortField = "d.updated_at"
		case "created_at":
			sortField = "d.created_at"
		case "document_type":
			sortField = "d.document_type"
		default:
			sortField = "d.created_at"
		}
	}

	if filters.SortOrder == "asc" {
		sortOrder = "ASC"
	}

	query += ` ORDER BY ` + sortField + ` ` + sortOrder

	// Add pagination
	query += ` LIMIT $` + fmt.Sprintf("%d", argIdx) + ` OFFSET $` + fmt.Sprintf("%d", argIdx+1)
	args = append(args, filters.Limit, filters.Offset)

	// Execute query
	rows := try.To1(q.Query(ctx, query, args...))
	defer rows.Close()

	results = make([]outbound.DocumentSearchResult, 0)

	for rows.Next() {
		var result outbound.DocumentSearchResult
		try.To(rows.Scan(
			&result.ID,
			&result.Name,
			&result.Description,
			&result.DocumentType,
			&result.Tags,
			&result.CreatedBy,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.LatestVersion,
			&result.LatestVersionStatus,
		))
		results = append(results, result)
	}

	try.To(rows.Err())

	return results, total, nil
}
