package grpc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/url"

	"github.com/dmad1989/urlcut/internal/config"
	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/dbstore"
	"github.com/dmad1989/urlcut/internal/jsonobject"
	pb "github.com/dmad1989/urlcut/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
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

	u, err := url.ParseRequestURI(req.Shorten)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Redirect: ParseRequestURI: %s", err.Error())
	}

	if u.Path == "" || u.Path == "/" {
		return nil, status.Error(codes.InvalidArgument, "Redirect: no path in url")
	}

	redirectURL, err := h.cutter.GetKeyByValue(ctx, u.Path)
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
	return &emptypb.Empty{}, status.Errorf(codes.OK, "ping ok")
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
	resBatch := make([]*pb.URLItem, len(req.Batch))

	for _, b := range batchResponse {
		resBatch = append(resBatch, &pb.URLItem{CorrelationId: b.ID, OriginalUrl: b.OriginalURL, ShortUrl: fmt.Sprintf("%s/%s", h.config.GetShortAddress(), b.ShortURL)})
	}

	return &pb.CutterJsonBatchResponse{Batch: resBatch}, nil
}
func (h *Handlers) UserUrls(ctx context.Context, req *emptypb.Empty) (*pb.UserUrlsResponse, error) {
	err, _ := ctx.Value(config.ErrorCtxKey).(error)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "UserUrls: not authorized: %s", err.Error())
	}

	urls, err := h.cutter.GetUserURLs(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, status.Error(codes.NotFound, "UserUrls: no urls found")
		}
		return nil, status.Errorf(codes.Internal, "UserUrls: getting all urls: %s", err.Error())
	}

	if len(urls) == 0 {
		return nil, status.Error(codes.NotFound, "UserUrls: no urls found")
	}

	for i := 0; i < len(urls); i++ {
		urls[i].ShortURL = fmt.Sprintf("%s/%s", h.config.GetShortAddress(), urls[i].ShortURL)
	}

	resBatch := make([]*pb.URLItem, len(urls))
	for _, b := range urls {
		resBatch = append(resBatch, &pb.URLItem{CorrelationId: b.ID, OriginalUrl: b.OriginalURL, ShortUrl: fmt.Sprintf("%s/%s", h.config.GetShortAddress(), b.ShortURL)})
	}

	return &pb.UserUrlsResponse{Batch: resBatch}, nil
}
func (h *Handlers) DeleteUsersUrls(ctx context.Context, req *pb.DeleteUserUrlsRequest) (*emptypb.Empty, error) {
	err, _ := ctx.Value(config.ErrorCtxKey).(error)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "UserUrls: not authorized: %s", err.Error())
	}

	var ids jsonobject.ShortIds
	user := ctx.Value(config.UserCtxKey)
	if user == nil {
		return nil, status.Error(codes.Unauthenticated, "DeleteUsersUrls: no user in context")
	}

	userID, ok := user.(string)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "DeleteUsersUrls: wrong user type in context")
	}

	go h.cutter.DeleteUrls(userID, ids)

	return &emptypb.Empty{}, status.Errorf(codes.OK, "")
}
func (h *Handlers) Stats(ctx context.Context, req *emptypb.Empty) (*pb.StatsResponse, error) {
	p, ok := peer.FromContext(ctx)
	if !ok || p.Addr.String() == "" {
		return nil, status.Error(codes.PermissionDenied, "Stats: peer from context not ok")
	}

	ip := net.ParseIP(p.Addr.String())
	if ip == nil {
		return nil, status.Error(codes.PermissionDenied, "Stats: parse peer addr")
	}
	cAddr, _, err := net.ParseCIDR(h.config.GetTrustedSubnet())
	if err != nil {
		return nil, status.Errorf(codes.PermissionDenied, "Stats: ParseCIDR: %s", err.Error())
	}
	if !cAddr.Equal(ip) {
		return nil, status.Error(codes.PermissionDenied, "no access for your IP")
	}
	stats, err := h.cutter.GetStats(ctx)

	if err != nil {
		return nil, status.Errorf(codes.Internal, "Stats: %s", err.Error())
	}
	return &pb.StatsResponse{Urls: int32(stats.URLs), Users: int32(stats.Users)}, nil
}
