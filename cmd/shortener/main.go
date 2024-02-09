package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/dmad1989/urlcut/internal/config"
	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/dbstore"
	"github.com/dmad1989/urlcut/internal/logging"
	"github.com/dmad1989/urlcut/internal/serverapi"
	"github.com/dmad1989/urlcut/internal/store"
)

func main() {
	ctx := context.Background()
	err := logging.Initilize()
	if err != nil {
		panic(err)
	}
	defer logging.Log.Sync()
	conf := config.ParseConfig()

	var storage cutter.Store
	if conf.GetDBConnName() != "" {
		storage, err = dbstore.New(ctx, conf)
		if err != nil {
			panic(err)
		}
		defer storage.CloseDB()
	} else {
		storage, err = store.New(ctx, conf)
		if err != nil {
			panic(err)
		}
	}
	app := cutter.New(storage)
	server := serverapi.New(app, conf)
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	err = server.Run(ctx)
	if err != nil {
		panic(err)
	}
}
