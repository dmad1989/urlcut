package dbstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/dmad1989/urlcut/internal/jsonobject"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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
	ALTER TABLE IF EXISTS public.urls OWNER to postgres;
	CREATE INDEX short_url ON urls (short_url);
    CREATE INDEX original_url ON urls (original_url);
	ALTER TABLE public.urls ADD CONSTRAINT urls_original_unique UNIQUE (original_url);
	`
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

	timeout = time.Duration(time.Second * 10)
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

	err = res.Ping(ctx)
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
func (te *UniqueURLError) Unwrap() error {
	return te.Err
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
