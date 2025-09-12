package database

import (
	"context"
	"database/sql"

	"github.com/qiniu/zeroops/internal/config"
)

type Database struct {
	db     *sql.DB
	config *config.DatabaseConfig
}

func NewDatabase(cfg *config.DatabaseConfig) (*Database, error) {
	database := &Database{config: cfg}
	return database, nil
}

func (d *Database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

func (d *Database) DB() *sql.DB {
	return d.db
}

// BeginTx 开始事务
func (d *Database) BeginTx(ctx context.Context) (*sql.Tx, error) {
	return d.db.BeginTx(ctx, nil)
}

// QueryContext 执行查询
func (d *Database) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.db.QueryContext(ctx, query, args...)
}

// QueryRowContext 执行单行查询
func (d *Database) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.db.QueryRowContext(ctx, query, args...)
}

// ExecContext 执行操作
func (d *Database) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.db.ExecContext(ctx, query, args...)
}
