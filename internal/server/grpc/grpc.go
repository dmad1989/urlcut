package grpc

import (
	"context"

	"github.com/dmad1989/urlcut/internal/jsonobject"
	pb "github.com/dmad1989/urlcut/proto"
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

type Server struct {
	pb.UnimplementedUrlCutServer
}

func New(cutter ICutter, config Configer) *Server {
	// api := &Server{cutter: cutter, config: config, mux: chi.NewMux()}
	// api.initHandlers()
	return &Server{}
}
