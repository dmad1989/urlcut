package handlers

import (
	"fmt"
	"hash/crc32"
	"io"
	"net/http"
)

var urls urlMap

func init() {
	if len(urls) == 0 {
		urls = make(urlMap, 4)
	}
}

func Manage(res http.ResponseWriter, req *http.Request) {
	switch req.Method {
	case http.MethodPost:
		cutter(res, req)
	case http.MethodGet:
		redirect(res, req)
	default:
		res.WriteHeader(http.StatusBadRequest)
	}
}

func cutter(res http.ResponseWriter, req *http.Request) {
	fmt.Println("I m in cutter")
	if req.Method != http.MethodPost {
		res.WriteHeader(http.StatusBadRequest)
	}

	body, err := io.ReadAll(req.Body)
	fmt.Println("I m in cutter- 2, ", req.URL)
	if err != nil {
		res.WriteHeader(http.StatusBadRequest)
		res.Write([]byte(err.Error()))
		return
	}
	res.Header().Set("Content-Type", "text/plain")
	// res.Header().Set("Content-Length", "30")
	res.WriteHeader(http.StatusCreated)
	res.Write([]byte(fmt.Sprintf("http://%s%s%s\n", req.Host, req.URL.Path, doCut(body))))
}

func doCut(body []byte) (generated string) {
	if urls.has(string(body)) {
		generated = urls.Get(string(body))
		return
	}
	crc32Table := crc32.MakeTable(crc32.IEEE)
	generated = fmt.Sprint(crc32.Checksum(body, crc32Table))
	urls.add(string(body), generated)
	return
}

func redirect(res http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		res.WriteHeader(http.StatusBadRequest)
	}
	fmt.Println("I m in redirect- 1")
	// Эндпоинт с методом GET и путём /{id}, где id — идентификатор сокращённого URL (например, /EwHXdJfB). В случае успешной обработки запроса сервер возвращает ответ с кодом 307 и оригинальным URL в HTTP-заголовке Location.
	// Пример запроса к серверу:
	// GET /EwHXdJfB HTTP/1.1
	// Host: localhost:8080
	// Content-Type: text/plain
	// Пример ответа от сервера:
	// HTTP/1.1 307 Temporary Redirect
	// Location: https://practicum.yandex.ru/

	//TODO
	// проверка входящего параметра
	// поиск в сохраненных значениях URL для редиректа
	v := urlMap.getKey("fd")
	//построение ответа от сервиса
	res.Header().Set("content-type", "application/json")
	// устанавливаем код 200
	res.WriteHeader(http.StatusOK)
	// пишем тело ответа
	res.Write([]byte("HI"))
}

type urlMap map[string]string

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

func (u urlMap) getKey(value string) string {
	if len(u) == 0 {
		return ""
	}
	for key, val := range u {
		if val == value {
			return key
		}
	}

}
