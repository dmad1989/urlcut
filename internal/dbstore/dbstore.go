package dbstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type conf interface {
	GetFileStoreName() string
	GetDBConnName() string
}

type storage struct {
	rw sync.RWMutex
	db *sql.DB
}

func New(c conf) (*storage, error) {
	if c.GetDBConnName() == "" {
		return nil, errors.New("init db storage: conn name is empty")
	}
	db, err := sql.Open("pgx", c.GetDBConnName())
	if err != nil {
		return nil, fmt.Errorf("conncet to DB: %w", err)
	}
	db.SetConnMaxLifetime(0)
	db.SetMaxIdleConns(50)
	db.SetMaxOpenConns(50)

	res := storage{rw: sync.RWMutex{},
		db: db}

	return &res, nil
}

func (s *storage) Ping(ctx context.Context) error {
	s.rw.RLock()
	defer s.rw.RUnlock()
	err := s.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("fail to ping db: %w", err)
	}
	return nil
}

func (s *storage) CloseDB() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("failed to close db conn: %w", err)
	}
	return nil
}

func (s *storage) Get(ctx context.Context, key string) (string, error) {
	//todo
	return "", errors.New("not supported yet")
}

func (s *storage) Add(ctx context.Context, key, value string) error {
	//todo
	return errors.New("not supported yet")
}

func (s *storage) GetKey(ctx context.Context, value string) (string, error) {
	// todo
	return "", errors.New("not supported yet")
}
