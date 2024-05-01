// main корневой модуль сервиса.
// запускает создание контекста, инициализацию слоев приложения, сервер.
// Хранилище может быть двух типов: БД Postgres или json-файл.
// Тип зависит от конфигурации при вызове. См описание пакета Config
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

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
	defer func() {
		err = logging.Log.Sync()
		if err != nil {
			logging.Log.Fatalf("log.sync in main: %w", err)
		}
	}()

	conf := config.ParseConfig()

	storage, err := initStore(ctx, conf)
	if err != nil {
		logging.Log.Fatalf("initStore: %w", err)
	}
	defer func() {
		err = storage.CloseDB()
		if err != nil {
			logging.Log.Fatalf("storage.CloseDB in main: %w", err)
		}
	}()
	app := cutter.New(storage)
	server := serverapi.New(app, conf)
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	err = server.Run(ctx)
	if err != nil {
		panic(err)
	}
}

// initStore отвечает за инициализацию хранилища сокращений.
//
// хранилище может быть трех видов: БД postgres, json-файл или внутреннее хранилище (map)
func initStore(ctx context.Context, conf config.Config) (storage cutter.Store, err error) {
	if conf.GetDBConnName() != "" {
		storage, err = dbstore.New(ctx, conf)
		if err != nil {
			return nil, fmt.Errorf("db: %w", err)
		}
		return
	}
	storage, err = store.New(ctx, conf)
	if err != nil {
		return nil, fmt.Errorf("store: %w", err)
	}
	return
}
