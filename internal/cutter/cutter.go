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
)

type Store interface {
	GetShortURL(ctx context.Context, key string) (string, error)
	Add(ctx context.Context, original, short string) error
	GetOriginalURL(ctx context.Context, value string) (res string, err error)
	Ping(context.Context) error
	CloseDB() error
	UploadBatch(ctx context.Context, batch jsonobject.Batch) (jsonobject.Batch, error)
	GetUserURLs(ctx context.Context) (jsonobject.Batch, error)
	CheckIsUserURL(ctx context.Context, shortURL string) (bool, error)
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

func (a *App) DeleteUrls(ctx context.Context, ids jsonobject.ShortIds) {
	doneCh := make(chan struct{})
	defer close(doneCh)
	numWorkers := len(ids)
	if numWorkers > 10 {
		numWorkers = 10
	}
	idsCh := generator(doneCh, ids)

	channels := make([]chan string, numWorkers)

	for i := 0; i < numWorkers; i++ {
		addResultCh := a.checkUrls(ctx, doneCh, idsCh)
		channels[i] = addResultCh
	}
	// fanIn for transaction update
}

func generator(doneCh chan struct{}, ids jsonobject.ShortIds) chan string {
	inCh := make(chan string)
	go func() {
		defer close(inCh)
		for _, id := range ids {
			select {
			case <-doneCh:
				return
			case inCh <- id:
			}
		}
	}()
	return inCh
}

// проверяет автора сокращения
func (a *App) checkUrls(ctx context.Context, doneCh chan struct{}, idsCh chan string) chan string {
	resCh := make(chan string)
	go func() {
		defer close(resCh)

	loop:
		for {
			select {
			case <-doneCh:
				return
			case id, ok := <-idsCh:
				if !ok {
					break loop
				}
				ok, err := a.storage.CheckIsUserURL(ctx, id)
				if err != nil {
					logging.Log.Fatal(fmt.Errorf("check Urls: %w", err))
					return
				}
				if ok {
					resCh <- id
				}
			}
		}

		// for data := range idsCh {
		// 	ok, err := a.storage.CheckIsUserURL(ctx, data)
		// 	if err != nil {
		// 		fmt.Errorf("check Urls: %w", err)
		// 	}
		// 	if ok {
		// 		select {
		// 		case <-doneCh:
		// 			return
		// 		case resCh <- data:
		// 		}
		// 	}
		// }
	}()
	return resCh
}
