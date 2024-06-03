package grpc

import (
	"context"
	"errors"
	"fmt"

	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/dbstore"
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

func (h *Handlers) Cutter(ctx context.Context, req *pb.CutterRequest) (*pb.CutterResponse, error) {
	if req == nil || req.Url == "" {
		return nil, status.Error(codes.DataLoss, "Cutter: empty data request")
	}
	code, err := h.cutter.Cut(ctx, req.Url)
	var errResult error
	if err != nil {
		var uerr *cutter.UniqueURLError
		if !errors.As(err, &uerr) {
			return nil, status.Errorf(codes.Internal, "Cutter: getting code for url: %s", err.Error())
		}
		errResult = status.Errorf(codes.AlreadyExists, "such code already exists")
		code = uerr.Code
	}
	res := pb.CutterResponse{Result: fmt.Sprintf("%s/%s", h.config.GetShortAddress(), code)}
	return &res, errResult
}

func (h *Handlers) Redirect(ctx context.Context, req *pb.RedirectRequest) (*pb.RedirectResponse, error) {
	if req == nil || req.Shorten == "" {
		return nil, status.Error(codes.DataLoss, "Redirect: empty data request")
	}
	redirectURL, err := h.cutter.GetKeyByValue(ctx, req.Shorten)
	if err != nil {
		if errors.Is(err, dbstore.ErrDeletedURL) {
			return nil, status.Error(codes.NotFound, "url already deleted")
		}
		return nil, status.Errorf(codes.Internal, "redirect: fetching url fo redirect: %s", err.Error())
	}

	return &pb.RedirectResponse{Url: redirectURL}, nil
}

func (h *Handlers) Ping(ctx context.Context, req *emptypb.Empty) (*emptypb.Empty, error) {
	err := h.cutter.PingDB(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ping:  %s", err.Error())
	}
	return nil, status.Errorf(codes.OK, "ping ok")
}

func (h *Handlers) CutterJsonBatch(ctx context.Context, req *pb.CutterJsonBatchRequest) (*pb.CutterJsonBatchResponse, error) {
	if req == nil || req.Batch == nil || len(req.Batch) == 0 {
		return nil, status.Error(codes.DataLoss, "CutterJsonBatch: empty data request")
	}
	var batchRequest jsonobject.Batch

	for _, b := range req.Batch {
		batchRequest = append(batchRequest, jsonobject.BatchItem{ID: b.CorrelationId, OriginalURL: b.OriginalUrl, ShortURL: b.ShortUrl})
	}
	batchResponse, err := h.cutter.UploadBatch(ctx, batchRequest)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "CutterJsonBatch: getting code for url: %s", err.Error())
	}

	res := pb.CutterJsonBatchResponse{Batch: req.Batch}

	for _, b := range batchResponse {
		res.Batch = append(res.Batch, &pb.URLItem{CorrelationId: b.ID, OriginalUrl: b.OriginalURL, ShortUrl: fmt.Sprintf("%s/%s", h.config.GetShortAddress(), b.ShortURL)})
	}

	return &res, nil
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
