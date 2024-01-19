package store

import (
	"errors"
	"fmt"
	"sync"
)

type storage struct {
	rw     *sync.RWMutex
	urlMap map[string]string
}

func New() *storage {
	res := storage{urlMap: make(map[string]string, 2), rw: &sync.RWMutex{}}

	return &res
}

func (store storage) Get(key string) (string, error) {
	store.rw.RLock()
	generated, isFound := store.urlMap[key]
	store.rw.RUnlock()
	if !isFound {
		return "", fmt.Errorf("generated code for url %s is not found", key)
	}
	return generated, nil
}

func (store storage) Add(key, value string) {
	store.rw.Lock()
	store.urlMap[key] = value
	store.rw.Unlock()
}

func (store storage) GetKey(value string) (res string, err error) {
	store.rw.RLock()
	defer store.rw.RUnlock()
	if len(store.urlMap) == 0 {
		return "", errors.New("urlMap is empty")
	}
	for key, val := range store.urlMap {
		if val == value {
			return key, nil
		}
	}
	if res == "" {
		return "", fmt.Errorf("no data found in urlMap for value %s", value)
	}
	return
}
