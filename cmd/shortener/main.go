package main

import (
	"net/http"

	handlers "github.com/dmad1989/urlcut/internal/handlers"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc(`/`, handlers.Manage)
	err := http.ListenAndServe(`:8080`, mux)
	if err != nil {
		panic(err)
	}
}
