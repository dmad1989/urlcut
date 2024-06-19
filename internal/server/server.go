// Package server is pre server layer
// starts and stops http and gRPC server implemetations
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

// Servers contains http and grpc variables of server interface
type Servers struct {
	httpServer server
	grpcServer server
}

// New creates new instanse of Servers
// And creates new instanse of http and grpc servers
func New(cutter http.ICutter, config http.Configer, ctx context.Context) *Servers {
	return &Servers{
		http.New(cutter, config, ctx),
		grpc.New(cutter, config)}
}

// Serve func starts servers using Run func
// Waits till context signal recieved
// Gracefully stoppes servers using Stop
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
