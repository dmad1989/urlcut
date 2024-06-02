package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/dmad1989/urlcut/internal/jsonobject"
	"github.com/dmad1989/urlcut/internal/logging"
	"github.com/dmad1989/urlcut/internal/server/auth"
	pb "github.com/dmad1989/urlcut/proto"
	iauth "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	ilog "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"go.uber.org/zap"
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
	grpc *grpc.Server
	h    *Handlers
}

func New(cutter ICutter, config Configer) *Server {
	opts := []ilog.Option{
		ilog.WithLogOnEvents(ilog.StartCall, ilog.FinishCall),
	}
	return &Server{
		grpc: grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				ilog.UnaryServerInterceptor(InterceptorLogger(logging.Log), opts...),
			),
			grpc.UnaryInterceptor(iauth.UnaryServerInterceptor(auth.GRPC)),
		),
		h: &Handlers{cutter: cutter, config: config}}
}

func (s *Server) Run(ctx context.Context) error {
	pb.RegisterUrlCutServer(s.grpc, s.h)
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

// InterceptorLogger copied from https://github.com/grpc-ecosystem/go-grpc-middleware/blob/main/interceptors/logging/examples/zap/example_test.go
func InterceptorLogger(l *zap.SugaredLogger) ilog.Logger {
	return ilog.LoggerFunc(func(ctx context.Context, lvl ilog.Level, msg string, fields ...any) {
		f := make([]zap.Field, 0, len(fields)/2)

		for i := 0; i < len(fields); i += 2 {
			key := fields[i]
			value := fields[i+1]

			switch v := value.(type) {
			case string:
				f = append(f, zap.String(key.(string), v))
			case int:
				f = append(f, zap.Int(key.(string), v))
			case bool:
				f = append(f, zap.Bool(key.(string), v))
			default:
				f = append(f, zap.Any(key.(string), v))
			}
		}

		logger := l.WithOptions(zap.AddCallerSkip(1)).With(f)

		switch lvl {
		case ilog.LevelDebug:
			logger.Debug(msg)
		case ilog.LevelInfo:
			logger.Info(msg)
		case ilog.LevelWarn:
			logger.Warn(msg)
		case ilog.LevelError:
			logger.Error(msg)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
