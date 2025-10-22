package db

import (
	"database/sql"
	"fmt"
	"nota/types"

	"net/url"
	"time"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/driver/pgdriver"
)

func NewDB(cfg types.AppConfig) (*bun.DB, error) {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		url.QueryEscape(cfg.DBUser),
		url.QueryEscape(cfg.DBPassword),
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBName,
	)

	sqlDb := sql.OpenDB(pgdriver.NewConnector(
		pgdriver.WithDSN(dbUrl),
	))
	sqlDb.SetMaxOpenConns(25)                 // Maximum open connections
	sqlDb.SetMaxIdleConns(10)                 // Maximum idle connections
	sqlDb.SetConnMaxLifetime(5 * time.Minute) // Connection lifetime
	sqlDb.SetConnMaxIdleTime(5 * time.Minute) // Idle connection timeout

	// Test the connection
	if err := sqlDb.Ping(); err != nil {
		return nil, err
	}

	db := bun.NewDB(sqlDb, pgdialect.New())

	return db, nil

}
