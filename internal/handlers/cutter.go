package handlers

import "net/http"

func CutterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		// write response 400
	}

	// 	Сервер принимает в теле запроса строку URL как text/plain и возвращает ответ с кодом 201 и сокращённым URL как text/plain.
	// Пример запроса к серверу:
	// POST / HTTP/1.1
	// Host: localhost:8080
	// Content-Type: text/plain

	// https://practicum.yandex.ru/

	// Пример ответа от сервера:
	// HTTP/1.1 201 Created
	// Content-Type: text/plain
	// Content-Length: 30

	// http://localhost:8080/EwHXdJfB

	//TODO
	// парсинг запроса
	// присвоение сокращения к ссылке - запись в место чтобы get мог использовать
	// построение ответа

}
