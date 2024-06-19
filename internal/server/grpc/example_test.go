package grpc

import (
	"context"
	"log"
	"strings"
	"time"

	"github.com/dmad1989/urlcut/internal/logging"
	pb "github.com/dmad1989/urlcut/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Example shows how to make client side of grpc server
// How to deal with token auth (if you dont have token it will be generated and returned to you)
func Example() {
	conn, err := grpc.Dial(":3200", grpc.WithTransportCredentials(insecure.NewCredentials()))
	defer logging.Log.Sync()
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := pb.NewUrlCutClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	md := metadata.New(map[string]string{"authorization": "berear  "})
	ctx = metadata.NewOutgoingContext(ctx, md)
	ctx = clientCutter(ctx, "http://ty.ry", c)
	clientCutter(ctx, "http://ta.ry", c)
}

func clientCutter(ctx context.Context, url string, c pb.UrlCutClient) context.Context {
	request := &pb.CutterRequest{Url: url}
	md, found := metadata.FromOutgoingContext(ctx)
	if !found || md.Get("authorization") == nil || len(md.Get("authorization")) == 0 {
		log.Fatal("md authorization not found")
	}
	var auth metadata.MD
	res, err := c.Cutter(ctx, request, grpc.Header(&auth))
	if err != nil {
		log.Fatal(err)
	}

	a := auth["authorization"]
	if len(a) == 0 {
		log.Fatalln("missing 'authorization' header")
	}
	if strings.Trim(a[0], " ") == "" {
		log.Fatalln("empty 'authorization' header")
	}

	logging.Log.Infof("cutter result", zap.String("result", res.Result), zap.String("token", a[0]))

	if md.Get("authorization")[0] == "berear  " {
		md.Set("authorization", a[0])
		return metadata.NewOutgoingContext(ctx, md)
	}
	return ctx
}
