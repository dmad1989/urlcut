package main

import (
	"github.com/dmad1989/urlcut/internal/config"
	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/serverapi"
	"github.com/dmad1989/urlcut/internal/store"
)

func main() {
	conf := config.ParseConfig()
	storage := store.New()
	app := cutter.New(storage)
	server := serverapi.New(app, conf)
	server.Run()
}
