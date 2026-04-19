package postgres

import (
	"context"
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/approval"
	"github.com/bbridges_11/document-registry/internal/domain/stakeholder"
	dbPostgres "github.com/bbridges_11/document-registry/internal/platform/database/postgres"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type StakeholderRepository struct {
	runner *dbPostgres.Runner
}

func NewStakeholderRepository(runner *dbPostgres.Runner) *StakeholderRepository {
	return &StakeholderRepository{runner: runner}
}

func (r *StakeholderRepository) Save(ctx context.Context, sh *stakeholder.Stakeholder) (err error) {
	defer err2.Handle(&err)
	q := r.runner.GetQuerier(ctx)
	query := `INSERT INTO document_stakeholders (document_id, user_id, role, created_at) VALUES ($1, $2, $3, $4)`
	try.To1(q.Exec(ctx, query, sh.DocumentID(), sh.UserID(), sh.Role(), sh.CreatedAt()))
	return nil
}

func (r *StakeholderRepository) GetByDocumentIDAndUserID(ctx context.Context, documentID uuid.UUID, userID string) (_ *stakeholder.Stakeholder, err error) {
	defer err2.Handle(&err)
	q := r.runner.GetQuerier(ctx)
	query := `SELECT document_id, user_id, role, created_at FROM document_stakeholders WHERE document_id = $1 AND user_id = $2`
	var docID uuid.UUID
	var uid string
	var role stakeholder.Role
	var createdAt time.Time
	err = q.QueryRow(ctx, query, documentID, userID).Scan(&docID, &uid, &role, &createdAt)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	try.To(err)
	return stakeholder.RehydrateStakeholder(docID, uid, role, createdAt), nil
}

func (r *StakeholderRepository) ListByDocumentID(ctx context.Context, documentID uuid.UUID) (shs []*stakeholder.Stakeholder, err error) {
	defer err2.Handle(&err)
	q := r.runner.GetQuerier(ctx)
	query := `SELECT document_id, user_id, role, created_at FROM document_stakeholders WHERE document_id = $1`
	rows := try.To1(q.Query(ctx, query, documentID))
	defer rows.Close()
	shs = make([]*stakeholder.Stakeholder, 0)
	for rows.Next() {
		var docID uuid.UUID
		var uid string
		var role stakeholder.Role
		var createdAt time.Time
		try.To(rows.Scan(&docID, &uid, &role, &createdAt))
		shs = append(shs, stakeholder.RehydrateStakeholder(docID, uid, role, createdAt))
	}
	return shs, nil
}

func (r *StakeholderRepository) Delete(ctx context.Context, documentID uuid.UUID, userID string) (err error) {
	defer err2.Handle(&err)
	q := r.runner.GetQuerier(ctx)
	query := `DELETE FROM document_stakeholders WHERE document_id = $1 AND user_id = $2`
	try.To1(q.Exec(ctx, query, documentID, userID))
	return nil
}

func (r *StakeholderRepository) ExistsByDocumentIDAndUserID(ctx context.Context, documentID uuid.UUID, userID string) (exists bool, err error) {
	defer err2.Handle(&err)
	q := r.runner.GetQuerier(ctx)
	query := `SELECT EXISTS(SELECT 1 FROM document_stakeholders WHERE document_id = $1 AND user_id = $2)`
	try.To(q.QueryRow(ctx, query, documentID, userID).Scan(&exists))
	return exists, nil
}

type ApprovalRepository struct {
	runner *dbPostgres.Runner
}

func NewApprovalRepository(runner *dbPostgres.Runner) *ApprovalRepository {
	return &ApprovalRepository{runner: runner}
}

func (r *ApprovalRepository) Save(ctx context.Context, appr *approval.Approval) (err error) {
	defer err2.Handle(&err)
	q := r.runner.GetQuerier(ctx)
	query := `INSERT INTO approvals (id, version_id, user_id, role, approved, comment, created_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`
	try.To1(q.Exec(ctx, query, appr.ID(), appr.VersionID(), appr.UserID(), appr.Role(), appr.Approved(), appr.Comment(), appr.CreatedAt()))
	return nil
}

func (r *ApprovalRepository) GetByID(ctx context.Context, id uuid.UUID) (_ *approval.Approval, err error) {
	defer err2.Handle(&err)
	q := r.runner.GetQuerier(ctx)
	query := `SELECT id, version_id, user_id, role, approved, comment, created_at FROM approvals WHERE id = $1`
	var apprID, versionID uuid.UUID
	var userID string
	var role approval.ApprovalRole
	var approved bool
	var comment string
	var createdAt time.Time
	err = q.QueryRow(ctx, query, id).Scan(&apprID, &versionID, &userID, &role, &approved, &comment, &createdAt)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	try.To(err)
	return approval.RehydrateApproval(apprID, versionID, userID, role, approved, comment, createdAt), nil
}

func (r *ApprovalRepository) ListByVersionID(ctx context.Context, versionID uuid.UUID) (apprs []*approval.Approval, err error) {
	defer err2.Handle(&err)
	q := r.runner.GetQuerier(ctx)
	query := `SELECT id, version_id, user_id, role, approved, comment, created_at FROM approvals WHERE version_id = $1 ORDER BY created_at DESC`
	rows := try.To1(q.Query(ctx, query, versionID))
	defer rows.Close()
	apprs = make([]*approval.Approval, 0)
	for rows.Next() {
		var apprID, verID uuid.UUID
		var userID string
		var role approval.ApprovalRole
		var approved bool
		var comment string
		var createdAt time.Time
		try.To(rows.Scan(&apprID, &verID, &userID, &role, &approved, &comment, &createdAt))
		apprs = append(apprs, approval.RehydrateApproval(apprID, verID, userID, role, approved, comment, createdAt))
	}
	return apprs, nil
}

func (r *ApprovalRepository) GetByVersionIDAndUserID(ctx context.Context, versionID uuid.UUID, userID string) (_ *approval.Approval, err error) {
	defer err2.Handle(&err)
	q := r.runner.GetQuerier(ctx)
	query := `SELECT id, version_id, user_id, role, approved, comment, created_at FROM approvals WHERE version_id = $1 AND user_id = $2`
	var apprID, verID uuid.UUID
	var uid string
	var role approval.ApprovalRole
	var approved bool
	var comment string
	var createdAt time.Time
	err = q.QueryRow(ctx, query, versionID, userID).Scan(&apprID, &verID, &uid, &role, &approved, &comment, &createdAt)
	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	try.To(err)
	return approval.RehydrateApproval(apprID, verID, uid, role, approved, comment, createdAt), nil
}

func (r *ApprovalRepository) ExistsByVersionIDAndUserID(ctx context.Context, versionID uuid.UUID, userID string) (exists bool, err error) {
	defer err2.Handle(&err)
	q := r.runner.GetQuerier(ctx)
	query := `SELECT EXISTS(SELECT 1 FROM approvals WHERE version_id = $1 AND user_id = $2)`
	try.To(q.QueryRow(ctx, query, versionID, userID).Scan(&exists))
	return exists, nil
}
