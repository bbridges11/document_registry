package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type txKey struct{}

type Querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Runner struct {
	pool *pgxpool.Pool
}

func NewRunner(pool *pgxpool.Pool) *Runner {
	return &Runner{pool: pool}
}

func (r *Runner) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) (err error) {
	defer err2.Handle(&err)

	tx := try.To1(r.pool.BeginTx(ctx, pgx.TxOptions{}))

	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
			return
		}
		err = tx.Commit(ctx)
	}()

	ctx = context.WithValue(ctx, txKey{}, tx)
	return fn(ctx)
}

func (r *Runner) GetQuerier(ctx context.Context) Querier {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return r.pool
}

func (r *Runner) Pool() *pgxpool.Pool {
	return r.pool
}
