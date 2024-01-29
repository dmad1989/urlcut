package cutter

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type store interface {
	Get(key string) (string, error)
	Add(key, value string) error
	GetKey(value string) (res string, err error)
}

type App struct {
	storage store
}

func New(s store) *App {
	return &App{storage: s}
}

func (a *App) Cut(url string) (generated string, err error) {
	generated, err = a.storage.Get(url)
	if err != nil {
		return "", fmt.Errorf("Cut: getting value by key %s from storage : %w", url, err)
	}

	if generated != "" {
		return
	}
	generated, err = randStringBytes(8)
	if err != nil {
		return "", fmt.Errorf("Cut: while generating path: %w", err)
	}
	err = a.storage.Add(url, generated)
	if err != nil {
		return "", fmt.Errorf("Cut: failed to add path: %w", err)
	}
	return
}

func (a *App) GetKeyByValue(value string) (res string, err error) {
	res, err = a.storage.GetKey(value)
	if err != nil {
		return "", fmt.Errorf("GetKeyByValue: while getting value by key:%s: %w", value, err)
	}
	return
}

func randStringBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("randStringBytes: Generating random string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
