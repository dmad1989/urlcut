package store

import (
	"errors"
	"fmt"
)

type urlMap map[string]string

func New() *urlMap {
	res := make(urlMap, 2)
	return &res
}

func (u urlMap) Get(key string) (string, error) {
	generated, isFound := u[key]
	if !isFound {
		return "", fmt.Errorf("generated code for url %s is not found", key)
	}
	return generated, nil
}

func (u urlMap) Add(key, value string) {
	u[key] = value
}

func (u urlMap) GetKey(value string) (res string, err error) {
	if len(u) == 0 {
		return "", errors.New("urlMap is empty")
	}
	for key, val := range u {
		if val == value {
			return key, nil
		}
	}
	if res == "" {
		return "", fmt.Errorf("no data found in urlMap for value %s", value)
	}
	return
}
