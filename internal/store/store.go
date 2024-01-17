package store

import "errors"

type urlMap map[string]string

func New() *urlMap {
	res := make(urlMap, 2)
	return &res
}

func (u urlMap) Get(key string) (string, error) {
	generated, isFound := u[key]
	if !isFound {
		return "", errors.New("generated code for url not found")
	}
	return generated, nil
}

func (u urlMap) Add(key, value string) {
	u[key] = value
}

func (u urlMap) GetKey(value string) (res string) {
	if len(u) == 0 {
		return
	}
	for key, val := range u {
		if val == value {
			return key
		}
	}
	return
}
