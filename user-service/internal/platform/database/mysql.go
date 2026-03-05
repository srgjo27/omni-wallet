package database

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	// MySQL driver — registered as a side-effect import.
	_ "github.com/go-sql-driver/mysql"
)

// NewMySQLConnection creates and validates a connection pool to MySQL.
// Connection pooling is configured to support high-concurrency workloads.
func NewMySQLConnection(dsn string, maxOpen, maxIdle int, connLifetime time.Duration) (*sqlx.DB, error) {
	db, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to mysql: %w", err)
	}

	// Configure the connection pool. These settings prevent resource exhaustion
	// under high TPS and ensure stale connections are recycled.
	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)
	db.SetConnMaxLifetime(connLifetime)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("pinging mysql: %w", err)
	}

	return db, nil
}
