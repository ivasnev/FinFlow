package db

import (
	"context"
	"database/sql"
	"fmt"
	"golang.org/x/xerrors"

	"gorm.io/gorm"
)

type TxContextKey struct{}

func GetTx(ctx context.Context, db *gorm.DB) *gorm.DB {
	tx, ok := ctx.Value(TxContextKey{}).(*gorm.DB)
	if !ok {
		return db
	}
	return tx
}

// WithTx - метод поднимает транзакцию с дефолтным уровнем транзакции postgress (Read Committed)
// и передает в контекст вложенной функции.
func WithTx(ctx context.Context, db *gorm.DB, txFunc func(ctx context.Context) error) (err error) {
	return WithTxIsolation(ctx, db, sql.LevelDefault, txFunc)
}

// WithTxIsolation - метод поднимает транзакцию и передает в контекст вложенной функции
// Данный метод помогает забирать транзакцию базы данных без передачи явной транзакции.
// Метод ExtractConn помогает забрать из контекста транзакцию.
func WithTxIsolation(
	ctx context.Context,
	db *gorm.DB,
	isolation sql.IsolationLevel,
	txFunc func(ctx context.Context) error,
) (err error) {
	tx := db.Begin(&sql.TxOptions{Isolation: isolation})
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			_ = tx.Rollback()
			panic(fmt.Sprintf("panic transation: %v", recoverErr))
		}
	}()

	ctx = context.WithValue(ctx, TxContextKey{}, tx)
	txFuncErr := txFunc(ctx)
	if txFuncErr != nil {
		_ = tx.Rollback()
		return txFuncErr
	}

	if commitErr := tx.Commit(); commitErr.Error != nil {
		return xerrors.Errorf("commit transaction error: `%w`", commitErr)
	}

	return nil
}
