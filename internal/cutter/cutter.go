package cutter

import (
	"crypto/rand"
	"encoding/base64"
)

type StoreMap interface {
	Get(key string) (string, error)
	Add(key, value string)
	GetKey(value string) (res string)
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
		return
	}
	app.urlsMap.Add(url, generated)
	return
}

func (app *App) GetKeyByValue(value string) string {
	return app.urlsMap.GetKey(value)
}

func randStringBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:n], nil
}
