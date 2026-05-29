package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vituu69/tiketis/postgres/sqlc"
)

// O armazenamento encapsula as consultas geradas e os auxiliares de transação.
type Store struct {
	*sqlc.Queries
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{
		Queries: sqlc.New(db),
		db:      db,
	}
}

// reporta os erros de serialização
func isSerializationError(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "40001"
}

// A função ExecTx executa a função fn dentro de uma transação e lida com o rollback em caso de erro.
// As falhas de serialização (SQLSTATE 40001) são repetidas automaticamente até o número máximo de tentativas definido em maxAttempts.
// feito para ajudar meu docker entrypoint
func (store *Store) ExecTx(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	const maxAttempts = 10
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		// roda uma serialização em attempt.
		lastErr = store.execTxOnce(ctx, fn)
		if lastErr == nil {
			return nil
		}
		if !isSerializationError(lastErr) {
			return lastErr
		}
		if attempt < maxAttempts-1 {
			if waitErr := sleepWithContext(ctx, retryWait(attempt)); waitErr != nil {
				return waitErr
			}
		}
	}
	return fmt.Errorf("transação falhou apos %d o attempts por conflito de serialização %w", maxAttempts, lastErr)
}

func (store *Store) execTxOnce(ctx context.Context, fn func(q *sqlc.Queries) error) error {
	// Use isolamento serializável para proteger fluxos que alteram o balanceamento de anomalias de condição de corrida.
	tx, err := store.db.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return err
	}

	defer tx.Rollback(ctx)

	// Bind sqlc queries para essa transação handle
	q := sqlc.New(tx)
	if err := fn(q); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit falhou %w", err)
	}
	return nil
}

func retryWait(attempt int) time.Duration {
	base := 50 * time.Millisecond
	for i := 0; i < attempt; i++ {
		base *= 2
		if base >= time.Second {
			return time.Second
		}
	}
	return base
}

func sleepWithContext(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}
