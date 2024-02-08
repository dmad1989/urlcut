package main

import (
	"github.com/dmad1989/urlcut/internal/config"
	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/logging"
	"github.com/dmad1989/urlcut/internal/serverapi"
	"github.com/dmad1989/urlcut/internal/store"
)

func main() {
	err := logging.Initilize()
	if err != nil {
		panic(err)
	}
	conf := config.ParseConfig()
	storage, err := store.New(conf)
	if err != nil {
		panic(err)
	}
	if storage.Database != nil {
		defer storage.Database.CloseDB()
	}
	app := cutter.New(storage)
	server := serverapi.New(app, conf)
	err = server.Run()
	if err != nil {
		panic(err)
	}
}
