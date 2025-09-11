package database

import (
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
	return d.db.Close()
}

func (d *Database) DB() *sql.DB {
	return d.db
}
