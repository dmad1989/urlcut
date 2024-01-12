package main

import (
	"github.com/dmad1989/urlcut/internal/config"
	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/serverapi"
	"github.com/dmad1989/urlcut/internal/store"
)

func main() {
	config.InitConfig()
	storage := store.New()
	cut := cutter.New(storage)
	server := serverapi.New(cut)
	server.Run()
}
