package urls

import (
	"crypto/rand"
	"encoding/base64"
)

func init() {
	if len(urls) == 0 {
		urls = make(urlMap, 4)
	}
}

type urlMap map[string]string

var urls urlMap

func GetUrls() urlMap {
	return urls
}

func Cut(body []byte) (generated string) {
	strBody := string(body)
	if urls.has(strBody) {
		generated = urls.Get(strBody)
		return
	}
	generated, err := randStringBytes(8)
	if err != nil {
		return
	}
	urls.add(strBody, generated)
	return
}

func (u urlMap) Get(key string) string {
	vs := u[key]
	if len(vs) == 0 {
		return ""
	}
	return vs
}

func (u urlMap) add(key, value string) {
	u[key] = value
}

func (u urlMap) has(key string) bool {
	_, ok := u[key]
	return ok
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

func randStringBytes(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b)[:n], nil
}
