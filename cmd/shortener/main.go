// main корневой модуль сервиса.
// запускает создание контекста, инициализацию слоев приложения, сервер.
// Хранилище может быть двух типов: БД Postgres или json-файл.
// Тип зависит от конфигурации при вызове. См описание пакета Config
//
// Для отображения информации о приложении при запуске нужно  указывать -ldflags:
// Build: -X 'main.buildVersion=${git describe --tags}'
// Commit: -X 'main.buildCommit=$(git rev-parse HEAD)'
// Date: -X 'main.buildDate=$(git show -s --format=%ai)'
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

var (
	buildVersion string
	buildDate    string
	buildCommit  string
)

func main() {
	fmt.Printf("Build version: %s\n", checkEmptyParam(buildVersion))
	fmt.Printf("Build date: %s\n", checkEmptyParam(buildDate))
	fmt.Printf("Build commit: %s\n", checkEmptyParam(buildCommit))

	ctx := context.Background()
	err := logging.Initilize()
	if err != nil {
		panic(err)
	}
	defer logging.Log.Sync()

	conf, err := config.ParseConfig()
	if err != nil {
		logging.Log.Errorf("parseConfig: %w", err)
	}

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

func checkEmptyParam(param string) string {
	if param == "" {
		return "N/A"
	}
	return param
}
