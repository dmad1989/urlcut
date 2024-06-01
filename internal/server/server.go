package server

import (
	"context"
	"fmt"

	grpc "github.com/dmad1989/urlcut/internal/server/grpc"
	http "github.com/dmad1989/urlcut/internal/server/http"
)

type server interface {
	Run(context.Context) error
	Stop()
}

type Servers struct {
	httpServer server //*http.Server
	grpcServer server //*grpc.Server
}

func New(cutter http.ICutter, config http.Configer, ctx context.Context) *Servers {
	return &Servers{
		http.New(cutter, config, ctx),
		grpc.New(cutter, config)}
}

func (s *Servers) Serve(ctx context.Context) (err error) {
	if err = s.httpServer.Run(ctx); err != nil {
		return fmt.Errorf("http server run: %w", err)
	}

	if err = s.grpcServer.Run(ctx); err != nil {
		return fmt.Errorf("grpc server run: %w", err)
	}
	<-ctx.Done()
	s.httpServer.Stop()
	s.grpcServer.Stop()
	return
}

func (s *Servers) Stop() {

}
