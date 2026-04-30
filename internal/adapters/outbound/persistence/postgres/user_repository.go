package postgres

import (
	"context"
	"strings"
	"time"

	"github.com/bbridges_11/document-registry/internal/domain/user"
	dbPostgres "github.com/bbridges_11/document-registry/internal/platform/database/postgres"
	"github.com/bbridges_11/document-registry/pkg/errors"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type UserRepository struct {
	runner *dbPostgres.Runner
}

func NewUserRepository(runner *dbPostgres.Runner) *UserRepository {
	return &UserRepository{runner: runner}
}

func (r *UserRepository) Save(ctx context.Context, usr *user.User) (err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		INSERT INTO users (id, external_id, email, name, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			role = EXCLUDED.role,
			active = EXCLUDED.active,
			updated_at = EXCLUDED.updated_at
	`

	try.To1(q.Exec(ctx, query,
		usr.ID(),
		usr.ExternalID(),
		usr.Email(),
		usr.Name(),
		string(usr.Role()),
		usr.Active(),
		usr.CreatedAt(),
		usr.UpdatedAt(),
	))

	return nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (_ *user.User, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT id, external_id, email, name, role, active, created_at, updated_at
		FROM users
		WHERE id = $1
	`

	var userID uuid.UUID
	var externalID, email, name, role string
	var active bool
	var createdAt, updatedAt time.Time

	err = q.QueryRow(ctx, query, id).Scan(
		&userID, &externalID, &email, &name, &role, &active, &createdAt, &updatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	try.To(err)

	return user.RehydrateUser(
		userID,
		externalID,
		email,
		name,
		user.UserRole(role),
		active,
		createdAt,
		updatedAt,
	), nil
}

func (r *UserRepository) GetByExternalID(ctx context.Context, externalID string) (_ *user.User, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT id, external_id, email, name, role, active, created_at, updated_at
		FROM users
		WHERE external_id = $1
	`

	var userID uuid.UUID
	var extID, email, name, role string
	var active bool
	var createdAt, updatedAt time.Time

	err = q.QueryRow(ctx, query, strings.ToLower(externalID)).Scan(
		&userID, &extID, &email, &name, &role, &active, &createdAt, &updatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	try.To(err)

	return user.RehydrateUser(
		userID,
		extID,
		email,
		name,
		user.UserRole(role),
		active,
		createdAt,
		updatedAt,
	), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (_ *user.User, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT id, external_id, email, name, role, active, created_at, updated_at
		FROM users
		WHERE email = $1
	`

	var userID uuid.UUID
	var externalID, userEmail, name, role string
	var active bool
	var createdAt, updatedAt time.Time

	err = q.QueryRow(ctx, query, email).Scan(
		&userID, &externalID, &userEmail, &name, &role, &active, &createdAt, &updatedAt,
	)

	if err == pgx.ErrNoRows {
		return nil, errors.ErrNotFound
	}
	try.To(err)

	return user.RehydrateUser(
		userID,
		externalID,
		userEmail,
		name,
		user.UserRole(role),
		active,
		createdAt,
		updatedAt,
	), nil
}

func (r *UserRepository) ListAll(ctx context.Context) (users []*user.User, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `
		SELECT id, external_id, email, name, role, active, created_at, updated_at
		FROM users
		ORDER BY created_at DESC
	`

	rows := try.To1(q.Query(ctx, query))
	defer rows.Close()

	users = make([]*user.User, 0)

	for rows.Next() {
		var userID uuid.UUID
		var externalID, email, name, role string
		var active bool
		var createdAt, updatedAt time.Time

		try.To(rows.Scan(&userID, &externalID, &email, &name, &role, &active, &createdAt, &updatedAt))

		users = append(users, user.RehydrateUser(
			userID,
			externalID,
			email,
			name,
			user.UserRole(role),
			active,
			createdAt,
			updatedAt,
		))
	}

	return users, nil
}

func (r *UserRepository) Exists(ctx context.Context, id uuid.UUID) (exists bool, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1 AND active = true)`

	try.To(q.QueryRow(ctx, query, id).Scan(&exists))

	return exists, nil
}

func (r *UserRepository) ExistsByExternalID(ctx context.Context, externalID string) (exists bool, err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `SELECT EXISTS(SELECT 1 FROM users WHERE external_id = $1)`

	try.To(q.QueryRow(ctx, query, strings.ToLower(externalID)).Scan(&exists))

	return exists, nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) (err error) {
	defer err2.Handle(&err)

	q := r.runner.GetQuerier(ctx)

	query := `DELETE FROM users WHERE id = $1`

	try.To1(q.Exec(ctx, query, id))

	return nil
}
