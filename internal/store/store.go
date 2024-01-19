package store

import (
	"fmt"
	"sync"
)

type storage struct {
	rw        *sync.RWMutex
	urlMap    map[string]string
	revertMap map[string]string
}

func New() *storage {
	res := storage{
		urlMap:    make(map[string]string, 2),
		rw:        &sync.RWMutex{},
		revertMap: make(map[string]string, 2)}

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
	store.revertMap[value] = key
	store.rw.Unlock()
}

func (store storage) GetKey(value string) (string, error) {
	store.rw.RLock()
	res, isFound := store.revertMap[value]
	store.rw.RUnlock()
	if !isFound {
		return "", fmt.Errorf("no data found in urlMap for value %s", value)
	}
	return res, nil
}
