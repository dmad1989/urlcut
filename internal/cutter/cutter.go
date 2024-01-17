package cutter

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

type StoreMap interface {
	Get(key string) (string, error)
	Add(key, value string)
	GetKey(value string) (res string, err error)
}

type App struct {
	urlsMap StoreMap
}

func New(storeMap StoreMap) *App {
	return &App{urlsMap: storeMap}
}

func (app *App) Cut(url string) (generated string, err error) {
	generated, _ = app.urlsMap.Get(url)
	if generated != "" {
		return
	}
	generated, err = randStringBytes(8)
	if err != nil {
		return "", fmt.Errorf("Error in cut method: %w", err)
	}
	app.urlsMap.Add(url, generated)
	return
}

func (app *App) GetKeyByValue(value string) (res string, err error) {
	res, err = app.urlsMap.GetKey(value)
	if err != nil {
		return "", fmt.Errorf("Error in GetKeyByValue: %w", err)
	}
	return
}

func randStringBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("Error while generating random string: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b)[:n], nil
}
