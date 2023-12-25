package handlers

import "net/http"

func Redirect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		// write response 400
	}
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
	//построение ответа от сервиса
}
