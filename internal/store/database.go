package store

import (
	"context"
	"database/sql"
	"fmt"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type DbStore struct {
	db *sql.DB
}

func initDB(dbSourceName string) (DbStore, error) {
	db, err := sql.Open("pgx", dbSourceName)
	if err != nil {
		return DbStore{db: nil}, fmt.Errorf("failed to conncet to DB %s, cause by %w", dbSourceName, err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)
	return DbStore{db: db}, nil
}

func (s DbStore) Ping(ctx context.Context) error {
	if s.db == nil {
		return fmt.Errorf("db is nil")
	}

	err := s.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("fail to ping db: %w", err)
	}
	return nil
}

func (s DbStore) CloseDB() error {
	if s.db == nil {
		return fmt.Errorf("db is nil")
	}
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close db conn: %w", err)
	}
	return nil
}
