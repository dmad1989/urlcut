package server

import (
	"context"
	"fmt"

	"github.com/dmad1989/urlcut/internal/jsonobject"
	grpc "github.com/dmad1989/urlcut/internal/server/grpc"
	http "github.com/dmad1989/urlcut/internal/server/http"
)

// ICutter интерфейс слоя с бизнес логикой
type ICutter interface {
	Cut(cxt context.Context, url string) (generated string, err error)
	GetKeyByValue(cxt context.Context, value string) (res string, err error)
	PingDB(context.Context) error
	UploadBatch(ctx context.Context, batch jsonobject.Batch) (jsonobject.Batch, error)
	GetUserURLs(ctx context.Context) (jsonobject.Batch, error)
	DeleteUrls(userID string, ids jsonobject.ShortIds)
	GetStats(ctx context.Context) (jsonobject.Stats, error)
}

// Configer интерйфейс конфигураци
type Configer interface {
	GetURL() string
	GetShortAddress() string
	GetEnableHTTPS() bool
	GetTrustedSubnet() string
}

type Servers struct {
	httpServer *http.Server
	grpcServer *grpc.Server
}

func New(cutter ICutter, config Configer) *Servers {
	return &Servers{http.New(cutter, config), grpc.New(cutter, config)}
}

func (s *Servers) Run(ctx context.Context) (err error) {
	err = s.httpServer.Run(ctx)
	if err != nil {
		return fmt.Errorf("http server run: %w", err)
	}

	return
}
