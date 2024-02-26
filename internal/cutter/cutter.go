package cutter

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/dmad1989/urlcut/internal/jsonobject"
	"github.com/dmad1989/urlcut/internal/logging"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"
)

type Store interface {
	GetShortURL(ctx context.Context, key string) (string, error)
	Add(ctx context.Context, original, short string) error
	GetOriginalURL(ctx context.Context, value string) (res string, err error)
	Ping(context.Context) error
	CloseDB() error
	UploadBatch(ctx context.Context, batch jsonobject.Batch) (jsonobject.Batch, error)
	GetUserURLs(ctx context.Context) (jsonobject.Batch, error)
	CheckIsUserURL(ctx context.Context, userID string, shortURL string) (bool, error)
	DeleteURLs(ctx context.Context, idsChs ...chan string) error
}

type App struct {
	storage Store
}

func New(s Store) *App {
	return &App{storage: s}
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

func (a *App) Cut(ctx context.Context, url string) (short string, err error) {
	short, err = randStringBytes(8)
	if err != nil {
		return "", fmt.Errorf("cut: while generating path: %w", err)
	}
	err = a.storage.Add(ctx, url, short)
	if err != nil {
		var uniq UniqueURLError
		var pgErr *pgconn.PgError
		switch {
		case errors.Is(err, &uniq):
			return "", err
		case errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation:
			short, err = a.storage.GetShortURL(ctx, url)
			if err != nil {
				return "", fmt.Errorf("cut: getting value by key %s from storage : %w", url, err)
			}
			err = NewUniqueURLError(short, err)
		default:
			return "", fmt.Errorf("cut: add path: %w", err)
		}
	}
	return
}

func (a *App) GetKeyByValue(ctx context.Context, value string) (res string, err error) {
	res, err = a.storage.GetOriginalURL(ctx, value)
	if err != nil {
		return "", fmt.Errorf("getKeyByValue: while getting value by key:%s: %w", value, err)
	}
	return
}

func (a *App) PingDB(ctx context.Context) error {
	return a.storage.Ping(ctx)
}

func (a *App) UploadBatch(ctx context.Context, batch jsonobject.Batch) (jsonobject.Batch, error) {
	for i := 0; i < len(batch); i++ {
		short, err := randStringBytes(8)
		if err != nil {
			return batch, fmt.Errorf("uploadBatch: %w", err)
		}
		(batch)[i].ShortURL = short
	}
	batch, err := a.storage.UploadBatch(ctx, batch)
	if err != nil {
		return batch, fmt.Errorf("UploadBact: %w", err)
	}
	return batch, nil
}

func randStringBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("randStringBytes: Generating random string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (a *App) GetUserURLs(ctx context.Context) (jsonobject.Batch, error) {
	res, err := a.storage.GetUserURLs(ctx)
	if err != nil {
		return nil, fmt.Errorf("cutter: %w", err)
	}
	return res, nil
}

func (a *App) DeleteUrls(userID string, ids jsonobject.ShortIds) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	doneCh := make(chan struct{})
	defer close(doneCh)
	numWorkers := len(ids)
	if numWorkers > 10 {
		numWorkers = 10
	}
	idsCh := generator(ctx, ids)
	chs, err := a.checks(ctx, numWorkers, idsCh, userID)
	if err != nil {
		logging.Log.Error(fmt.Errorf("DeleteURLs: %w", err))
		cancel()
	}
	err = a.storage.DeleteURLs(ctx, chs...)
	if err != nil {
		logging.Log.Error(fmt.Errorf("DeleteURLs: %w", err))
		cancel()
	}
}

func generator(ctx context.Context, ids jsonobject.ShortIds) chan string {
	inCh := make(chan string)
	go func() {
		defer close(inCh)
		for _, id := range ids {
			select {
			case <-ctx.Done():
				return
			case inCh <- id:
			}
		}
	}()
	return inCh
}

func (a *App) checks(ctx context.Context, numWorkers int, idsCh chan string, userID string) ([]chan string, error) {
	chs := make([]chan string, numWorkers)

	for i := 0; i < numWorkers; i++ {
		ch, err := a.checkUrls(ctx, idsCh, userID)
		if err != nil {
			return nil, fmt.Errorf("cutter checks: %w", err)
		}
		chs[i] = ch
	}
	return chs, nil
}

// проверяет автора сокращения
func (a *App) checkUrls(ctx context.Context, idsCh chan string, userID string) (chan string, error) {
	nctx, cancel := context.WithCancel(ctx)
	resCh := make(chan string)
	go func() {
		defer close(resCh)

		for id := range idsCh {
			ok, err := a.storage.CheckIsUserURL(nctx, userID, id)
			if err != nil {
				logging.Log.Errorw("Error on checkUrls", zap.String("URL", id), zap.String("user", userID), zap.Error(err))
				cancel()
			}
			if ok {
				resCh <- id
			}
		}
	}()

	return resCh, nil
}
