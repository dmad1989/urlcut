package store

import (
	"fmt"
	"sync"
)

type storage struct {
	rw        sync.RWMutex
	urlMap    map[string]string
	revertMap map[string]string
}

func New() *storage {
	res := storage{
		urlMap:    make(map[string]string, 2),
		rw:        sync.RWMutex{},
		revertMap: make(map[string]string, 2)}

	return &res
}

func (s *storage) Get(key string) (string, error) {
	s.rw.RLock()
	generated, isFound := s.urlMap[key]
	s.rw.RUnlock()
	if !isFound {
		return "", fmt.Errorf("generated code for url %s is not found", key)
	}
	return generated, nil
}

func (s *storage) Add(key, value string) {
	s.rw.Lock()
	s.urlMap[key] = value
	s.revertMap[value] = key
	s.rw.Unlock()
}

func (s *storage) GetKey(value string) (string, error) {
	s.rw.RLock()
	res, isFound := s.revertMap[value]
	s.rw.RUnlock()
	if !isFound {
		return "", fmt.Errorf("no data found in urlMap for value %s", value)
	}
	return res, nil
}
