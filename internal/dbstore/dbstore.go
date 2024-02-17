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

	"github.com/dmad1989/urlcut/internal/jsonobject"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	case err == sql.ErrNoRows:
		return "", nil
	case err != nil:
		return "", fmt.Errorf("dbstore.GetShortURL select: %w", err)
	default:
		return sURL, nil
	}
}

type UniqueURLError struct {
	Code string
	Err  error
}

func (ue *UniqueURLError) Error() string {
	return fmt.Sprintf("URL is not unique. Saved Code is: %s; %v", ue.Code, ue.Err)
}
func NewUniqueURLError(code string, err error) error {
	return &UniqueURLError{
		Code: code,
		Err:  err,
	}
}
func (ue *UniqueURLError) Unwrap() error {
	return ue.Err
}

func (s *storage) Add(ctx context.Context, original, short string) error {
	s.rw.RLock()
	defer s.rw.RUnlock()
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, err := s.db.ExecContext(tctx, sqlInsert, short, original)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr); pgErr.Code == pgerrcode.UniqueViolation {
			sURL := ""
			errQuery := s.db.QueryRowContext(tctx, sqlGetShortURL, original).Scan(&sURL)
			if errQuery != nil {
				return fmt.Errorf("dbstore.GetShortURL select: %w", err)
			}
			return NewUniqueURLError(sURL, err)
		}
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

func (s *storage) UploadBatch(ctx context.Context, batch *jsonobject.Batch) (*jsonobject.Batch, error) {
	s.rw.RLock()
	defer s.rw.RUnlock()
	tctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	tx, err := s.db.BeginTx(tctx, nil)
	if err != nil {
		return batch, fmt.Errorf("upload batch, transation begin: %w", err)
	}
	defer tx.Commit()

	stmtInsert, err := tx.PrepareContext(tctx, sqlInsert)
	if err != nil {
		return batch, fmt.Errorf("upload batch,, prepare stmt: %w", err)
	}
	defer stmtInsert.Close()
	stmtCheck, err := s.db.PrepareContext(tctx, sqlGetShortURL)
	if err != nil {
		return batch, fmt.Errorf("upload batch,, prepare stmt: %w", err)
	}
	defer stmtCheck.Close()
	for i := 0; i < len(*batch); i++ {
		var dbOriginalURL string
		err := stmtCheck.QueryRowContext(tctx, (*batch)[i].OriginalURL).Scan(&dbOriginalURL)
		switch {
		case err == sql.ErrNoRows:
			if _, err = stmtInsert.ExecContext(tctx, (*batch)[i].ShortURL, (*batch)[i].OriginalURL); err != nil {
				tx.Rollback()
				return batch, fmt.Errorf("batch insert %w", err)
			}
		case err != nil:
			tx.Rollback()
			return batch, fmt.Errorf("batch check %w", err)
		default:
			(*batch)[i].ShortURL = dbOriginalURL
		}
		(*batch)[i].OriginalURL = ""
	}
	return batch, nil
}
