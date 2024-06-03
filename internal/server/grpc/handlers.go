package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/jsonobject"
	pb "github.com/dmad1989/urlcut/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Handlers struct {
	pb.UnimplementedUrlCutServer
	cutter ICutter
	config Configer
}

func (h *Handlers) CutterJson(ctx context.Context, req *pb.CutterRequest) (*pb.CutterResponse, error) {
	if req == nil || req.Url == "" {
		return nil, status.Error(codes.DataLoss, "cutterJson: empty data request")
	}
	reqJSON := jsonobject.Request{URL: req.Url}
	code, err := h.cutter.Cut(ctx, reqJSON.URL)
	var errResult error
	if err != nil {
		var uerr *cutter.UniqueURLError
		if !errors.As(err, &uerr) {
			return nil, status.Errorf(codes.Internal, "cutterJson: getting code for url: %s", err.Error())
		}
		errResult = status.Errorf(codes.AlreadyExists, "such code already exists")
		code = uerr.Code
	}
	res := pb.CutterResponse{Result: fmt.Sprintf("%s/%s", h.config.GetShortAddress(), code)}
	return &res, errResult
}

func (h *Handlers) Cutter(context.Context, *pb.CutterRequest) (*pb.CutterResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Cutter not implemented")
}
func (h *Handlers) Redirect(context.Context, *pb.RedirectRequest) (*pb.RedirectResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Redirect not implemented")
}
func (h *Handlers) Ping(context.Context, *emptypb.Empty) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Ping not implemented")
}
func (h *Handlers) CutterJsonBatch(context.Context, *pb.CutterJsonBatchRequest) (*pb.CutterJsonBatchResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method CutterJsonBatch not implemented")
}
func (h *Handlers) UserUrls(context.Context, *emptypb.Empty) (*pb.UserUrlsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method UserUrls not implemented")
}
func (h *Handlers) DeleteUsersUrls(context.Context, *pb.DeleteUserUrlsRequest) (*emptypb.Empty, error) {
	return nil, status.Errorf(codes.Unimplemented, "method DeleteUsersUrls not implemented")
}
func (h *Handlers) Stats(context.Context, *emptypb.Empty) (*pb.StatsResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Stats not implemented")
}
