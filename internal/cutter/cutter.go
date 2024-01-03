package cutter

import (
	"crypto/rand"
	"encoding/base64"
)

type StoreMap interface {
	Get(key string) string
	Add(key, value string)
	Has(key string) bool
	GetKey(value string) (res string)
}

type storage struct {
	urlsMap StoreMap
}

func New(storeMap StoreMap) *storage {
	return &storage{urlsMap: storeMap}
}

func (store *storage) Cut(body []byte) (generated string) {
	strBody := string(body)
	if store.urlsMap.Has(strBody) {
		generated = store.urlsMap.Get(strBody)
		return
	}
	generated, err := randStringBytes(8)
	if err != nil {
		return
	}
	store.urlsMap.Add(strBody, generated)
	return
}

func (store *storage) GetKeyByValue(value string) string {
	return store.urlsMap.GetKey(value)
}

func randStringBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:n], nil
}
