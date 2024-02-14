package cutter

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"

	"github.com/dmad1989/urlcut/internal/jsonobject"
)

type Store interface {
	GetShortURL(ctx context.Context, key string) (string, error)
	Add(ctx context.Context, original, short string) error
	GetOriginalURL(ctx context.Context, value string) (res string, err error)
	Ping(context.Context) error
	CloseDB() error
	UploadBatch(ctx context.Context, batch *jsonobject.Batch) (*jsonobject.Batch, error)
}

type App struct {
	storage Store
}

func New(s Store) *App {
	return &App{storage: s}
}

func (a *App) Cut(ctx context.Context, url string) (short string, err error) {
	// short, err = a.storage.GetShortURL(ctx, url)
	// if err != nil {
	// 	return "", fmt.Errorf("cut: getting value by key %s from storage : %w", url, err)
	// }

	// if short != "" {
	// 	return
	// }
	short, err = randStringBytes(8)
	if err != nil {
		return "", fmt.Errorf("cut: while generating path: %w", err)
	}
	err = a.storage.Add(ctx, url, short)
	if err != nil {
		return "", fmt.Errorf("cut: add path: %w", err)
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

func (a *App) UploadBatch(ctx context.Context, batch *jsonobject.Batch) (*jsonobject.Batch, error) {
	for i := 0; i < len(*batch); i++ {
		short, err := randStringBytes(8)
		if err != nil {
			return batch, fmt.Errorf("uploadBatch: %w", err)
		}
		(*batch)[i].ShortURL = short
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
