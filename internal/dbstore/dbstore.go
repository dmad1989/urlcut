package dbstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	sqlCreateTable = `CREATE TABLE IF NOT EXISTS public.urls
	(
		"ID" bigint NOT NULL GENERATED ALWAYS AS IDENTITY ( INCREMENT 1 START 1 MINVALUE 1 MAXVALUE 1000000 CACHE 1 ),
		short_url text COLLATE pg_catalog."default" NOT NULL,
		original_url text COLLATE pg_catalog."default",
		CONSTRAINT urls_pkey PRIMARY KEY ("ID")
	)	
	TABLESPACE pg_default;	
	ALTER TABLE IF EXISTS public.urls OWNER to postgres;`
	sqlChechTableExists = `SELECT EXISTS (
		SELECT FROM 
			information_schema.tables 
		WHERE 
			table_schema LIKE 'public' AND 
			table_type LIKE 'BASE TABLE' AND
			table_name = 'urls'
		)`
	sqlGetShortURL    = "select  u.short_url  from urls u where u.original_url  = $1"
	sqlGetOriginalURL = "select  u.original_url  from urls u where u.short_url  = $1"
	sqlInsert         = "INSERT INTO public.urls (short_url, original_url) VALUES( $1, $2)"
	timeout           = time.Duration(time.Second * 10)
)

// todo общее значения для контекста таймаута

type conf interface {
	GetFileStoreName() string
	GetDBConnName() string
}

type storage struct {
	rw sync.RWMutex
	db *sql.DB
}

func New(ctx context.Context, c conf) (*storage, error) {
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

	res.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("check DB after create: %w", err)
	}
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	row := db.QueryRowContext(tctx, sqlChechTableExists)
	var tableExists bool
	err = row.Scan(&tableExists)
	if err != nil {
		return nil, fmt.Errorf("check table exists: %w", err)
	}
	if !tableExists {
		_, err = db.ExecContext(tctx, sqlCreateTable)
		if err != nil {
			return nil, fmt.Errorf("create table: %w", err)
		}
	}

	return &res, nil
}

func (s *storage) Ping(ctx context.Context) error {
	s.rw.RLock()
	defer s.rw.RUnlock()
	err := s.db.PingContext(ctx)
	if err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	return nil
}

func (s *storage) CloseDB() error {
	if err := s.db.Close(); err != nil {
		return fmt.Errorf("close db conn: %w", err)
	}
	return nil
}

func (s *storage) GetShortURL(ctx context.Context, key string) (string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	sURL := ""
	err := s.db.QueryRowContext(tctx, sqlGetShortURL, key).Scan(&sURL)

	switch {
	case err == sql.ErrNoRows:
		return "", nil
	case err != nil:
		return "", fmt.Errorf("dbstore.GetShortURL select: %w", err)
	default:
		return sURL, nil
	}
}

func (s *storage) Add(ctx context.Context, original, short string) error {
	s.rw.RLock()
	defer s.rw.RUnlock()
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := s.db.ExecContext(tctx, sqlInsert, short, original)
	if err != nil {
		return fmt.Errorf("dbstore.add: write items: %w", err)
	}
	return nil
}

func (s *storage) GetOriginalURL(ctx context.Context, value string) (string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	sURL := ""
	err := s.db.QueryRowContext(tctx, sqlGetOriginalURL, value).Scan(&sURL)

	switch {
	case err == sql.ErrNoRows:
		return "", fmt.Errorf("no data found in db for value %s", value)
	case err != nil:
		return "", fmt.Errorf("dbstore.GetOriginalURL select: %w", err)
	default:
		return sURL, nil
	}
}
