package serverapi

import (
	"fmt"
	"strings"
)

func ExampleServer_cutterHandler() {
	//запуск серевера
	_, s := initEnv()
	// Вызываем эндпоинт для сокращения URL
	// Метод POST , путь /
	res, err := s.Client().Post(s.URL, "text/plain", strings.NewReader("http://ya.ru"))

	//Обрабатываем ошибку
	if err != nil {
		fmt.Println(fmt.Errorf("in request: %w", err))
	}
	// 201 - Сокращен успешно
	// 400 - ошибка
	fmt.Println(res.Status)
	res.Body.Close()
	res, err = s.Client().Post(s.URL, "text/plain", strings.NewReader("qwerty12345"))
	//Обрабатываем ошибку
	if err != nil {
		fmt.Println(fmt.Errorf("in request: %w", err))
	}
	// 201 - Сокращен успешно
	// 400 - ошибка
	fmt.Println(res.Status)
	res.Body.Close()

	// Output:
	// 201 Created
	// 400 Bad Request
}

func ExampleServer_redirectHandler() {
	//запуск серевера
	_, s := initEnv()
	// Вызываем эндпоинт для сокращения URL
	// Метод GET , путь / + код скоращения
	res, err := s.Client().Get(s.URL)
	//Обрабатываем ошибку
	if err != nil {
		fmt.Println(fmt.Errorf("in request: %w", err))
	}
	defer res.Body.Close()
	// 307 Переход по сокращенному URL
	// 401  Ошибка авторизации
	// 410 URL Удален
	// 400 Ошибка
	fmt.Println(res.Status)
}
