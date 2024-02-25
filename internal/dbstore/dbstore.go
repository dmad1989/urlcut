package dbstore

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/pressly/goose/v3"

	"github.com/dmad1989/urlcut/internal/config"
	"github.com/dmad1989/urlcut/internal/jsonobject"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const timeout = time.Duration(time.Second * 10)

//go:embed sql/migrations/00001_create_urls_table.sql
var embedMigrations embed.FS

//go:embed sql/checkTableExists.sql
var sqlCheckTableExists string

//go:embed sql/getShortURL.sql
var sqlGetShortURL string

//go:embed sql/getOriginalURL.sql
var sqlGetOriginalURL string

//go:embed sql/insertURL.sql
var sqlInsert string

//go:embed sql/getUrlsByAuthor.sql
var sqlGetUrlsByAuthor string

//go:embed sql/markDelete.sql
var sqlMarkDelete string

//go:embed sql/checkUserURLExists.sql
var sqlCheckUserURLExists string

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

	if err = res.Ping(ctx); err != nil {
		return nil, fmt.Errorf("check DB after create: %w", err)
	}
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	row := db.QueryRowContext(tctx, sqlCheckTableExists)
	if row.Err() != nil {
		return nil, fmt.Errorf("check table exists: %w", err)
	}
	var tableExists bool
	if err = row.Scan(&tableExists); err != nil {
		return nil, fmt.Errorf("check table exists: %w", err)
	}
	if !tableExists {
		goose.SetBaseFS(embedMigrations)

		if err := goose.SetDialect("postgres"); err != nil {
			return nil, fmt.Errorf("goose.SetDialect: %w", err)
		}

		if err := goose.Up(db, "sql/migrations"); err != nil {
			return nil, fmt.Errorf("goose: create table: %w", err)
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
	case errors.Is(err, sql.ErrNoRows):
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
	userID := ctx.Value(config.UserCtxKey)
	if _, err := s.db.ExecContext(tctx, sqlInsert, short, original, userID); err != nil {
		return fmt.Errorf("dbstore.add: write items: %w", err)
	}
	return nil
}

var ErrorDeletedURL = errors.New("url was deleted")

func (s *storage) GetOriginalURL(ctx context.Context, value string) (string, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	sURL := ""
	isDeleted := false
	err := s.db.QueryRowContext(tctx, sqlGetOriginalURL, value).Scan(&sURL, &isDeleted)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf("no data found in db for value %s", value)
	case err != nil:
		return "", fmt.Errorf("dbstore.GetOriginalURL select: %w", err)
	case isDeleted:
		return "", ErrorDeletedURL
	default:
		return sURL, nil
	}
}

func (s *storage) UploadBatch(ctx context.Context, batch jsonobject.Batch) (jsonobject.Batch, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	userID := ctx.Value(config.UserCtxKey)
	if userID == "" {
		return batch, errors.New("upload batch, no user in context")
	}
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	tx, err := s.db.BeginTx(tctx, nil)
	if err != nil {
		return batch, fmt.Errorf("upload batch, transation begin: %w", err)
	}
	defer tx.Commit()

	stmtInsert, err := tx.PrepareContext(tctx, sqlInsert)
	if err != nil {
		return batch, fmt.Errorf("upload batch, prepare stmt: %w", err)
	}
	defer stmtInsert.Close()
	stmtCheck, err := s.db.PrepareContext(tctx, sqlGetShortURL)
	if err != nil {
		return batch, fmt.Errorf("upload batch, prepare stmt: %w", err)
	}
	defer stmtCheck.Close()
	for i := 0; i < len(batch); i++ {
		var dbOriginalURL string
		err := stmtCheck.QueryRowContext(tctx, batch[i].OriginalURL).Scan(&dbOriginalURL)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			if _, err = stmtInsert.ExecContext(tctx, batch[i].ShortURL, batch[i].OriginalURL, userID); err != nil {
				tx.Rollback()
				return batch, fmt.Errorf("batch insert: %w", err)
			}
		case err != nil:
			tx.Rollback()
			return batch, fmt.Errorf("batch check: %w", err)
		default:
			batch[i].ShortURL = dbOriginalURL
		}
		batch[i].OriginalURL = ""
	}
	return batch, nil
}

func (s *storage) GetUserURLs(ctx context.Context) (jsonobject.Batch, error) {
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	var res jsonobject.Batch
	userID := ctx.Value(config.UserCtxKey)
	if userID == "" {
		return res, errors.New("GetUserUrls, no user in context")
	}

	stmt, err := s.db.PrepareContext(tctx, sqlGetUrlsByAuthor)
	if err != nil {
		return nil, fmt.Errorf("GetUserUrls, prepare stmt: %w", err)
	}
	defer stmt.Close()
	s.rw.RLock()
	defer s.rw.RUnlock()
	rows, err := stmt.QueryContext(tctx, userID)
	if err != nil {
		return nil, fmt.Errorf("GetUserUrls, QueryContext: %w", err)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("GetUserUrls, QueryContext: %w", rows.Err())
	}

	for rows.Next() {
		var original string
		var short string
		err = rows.Scan(&short, &original)
		if err != nil {
			return nil, fmt.Errorf("GetUserUrls, scan db results %w", err)
		}
		res = append(res, jsonobject.BatchItem{OriginalURL: original, ShortURL: short})
	}
	return res, nil
}

func (s *storage) CheckIsUserURL(ctx context.Context, shortURL string) (bool, error) {
	res := false
	userID := ctx.Value(config.UserCtxKey)
	if userID == "" {
		return res, errors.New("CheckIsUserURL, no user in context")
	}
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	s.rw.RLock()
	defer s.rw.RUnlock()
	err := s.db.QueryRowContext(tctx, sqlCheckUserURLExists, shortURL, userID).Scan(&res)

	if err != nil {
		return false, fmt.Errorf("dbstore.CheckIsUserURL : %w", err)
	}
	return res, nil
}
