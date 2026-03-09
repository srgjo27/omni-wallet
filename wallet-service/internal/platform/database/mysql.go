package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	_ "github.com/go-sql-driver/mysql"
)

type txKey struct{}

func NewMySQLConnection(dsn string, maxOpen, maxIdle int, connLifetime time.Duration) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to mysql: %w", err)
	}
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(connLifetime)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging mysql: %w", err)
	}
	return db, nil
}
