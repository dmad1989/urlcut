package main

import (
	"fmt"
	"net/http"

	handlers "github.com/dmad1989/urlcut/internal/handlers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handlers.Manage)
	err := http.ListenAndServe(`:8080`, mux)
	fmt.Println("main err:", err)
	if err != nil {
		panic(err)
	}
}
