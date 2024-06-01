package grpc

import (
	"context"
	"net"

	"github.com/dmad1989/urlcut/internal/jsonobject"
	"github.com/dmad1989/urlcut/internal/logging"
	pb "github.com/dmad1989/urlcut/proto"
	"google.golang.org/grpc"
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
	grpc *grpc.Server
}

func New(cutter ICutter, config Configer) *Server {
	return &Server{grpc: grpc.NewServer()}
}

func (s *Server) Run(ctx context.Context) error {
	// pb.RegisterUrlCutServer(grpcServer, &s)
	go func() {
		l, err := net.Listen("tcp", ":3200")
		if err != nil {
			logging.Log.Errorf("listen tcp port 3200 %w", err)
		}
		logging.Log.Info("gRPC server started")
		err = s.grpc.Serve(l)
		if err != nil {
			logging.Log.Errorf("grps server serve: %w", err)
		}
	}()

	return nil
}

func (s *Server) Stop() {
	logging.Log.Info("gRPC server closed")
	s.grpc.GracefulStop()
}
