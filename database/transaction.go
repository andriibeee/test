package database

import (
	"context"
	"database/sql"
)

type TransactionManager interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context)
	Rollback(ctx context.Context)
}

const TX_KEY = "tx"

func getTransaction(ctx context.Context) *sql.Tx {
	return *(ctx.Value(TX_KEY).(**sql.Tx))
}

type TransactionManagerImpl struct {
	db *sql.DB
}

func NewTransactionManager(db *sql.DB) TransactionManager {
	return TransactionManagerImpl{db: db}
}

func (tm TransactionManagerImpl) Begin(ctx context.Context) (context.Context, error) {
	tx, err := tm.db.Begin()
	ctx = context.WithValue(ctx, TX_KEY, &tx)
	return ctx, err
}

func (tm TransactionManagerImpl) Commit(ctx context.Context) {
	tx := getTransaction(ctx)
	if tx != nil {
		tx.Commit()
	}
}

func (tm TransactionManagerImpl) Rollback(ctx context.Context) {
	tx := getTransaction(ctx)
	if tx != nil {
		tx.Rollback()
	}
}
