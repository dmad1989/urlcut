package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStore struct {
	db *sql.DB
}

func initDB(dbSourceName string) (DBStore, error) {
	db, err := sql.Open("pgx", dbSourceName)
	if err != nil {
		return DBStore{db: nil}, fmt.Errorf("failed to conncet to DB %s, cause by %w", dbSourceName, err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)
	return DBStore{db: db}, nil
}

func (s DBStore) Ping(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("db is nil")
	}

	err := s.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("fail to ping db: %w", err)
	}
	return nil
}

func (s DBStore) CloseDB() error {
	if s.db == nil {
		return fmt.Errorf("db is nil")
	}
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close db conn: %w", err)
	}
	return nil
}
