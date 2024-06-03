package grpc

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/dmad1989/urlcut/internal/cutter"
	"github.com/dmad1989/urlcut/internal/mocks"
	pb "github.com/dmad1989/urlcut/proto"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defURL        = "http://localhost:8080"
	defURLPattern = "http://localhost:8080/%s"
	defCode       = "code"
)

var errorUnique cutter.UniqueURLError = cutter.UniqueURLError{Code: defCode}

func TestCutterJson(t *testing.T) {
	type mock struct {
		config *mocks.MockConfiger
		cutter *mocks.MockICutter
	}

	type want struct {
		response *pb.CutterResponse
		err      assert.ErrorAssertionFunc
		code     codes.Code
	}

	tests := []struct {
		name    string
		request *pb.CutterRequest
		pMocks  func(m *mock)
		w       want
	}{
		{
			name:    "empty request",
			request: &pb.CutterRequest{},
			pMocks: func(m *mock) {
				gomock.InOrder(
					m.cutter.EXPECT().Cut(gomock.Any(), gomock.Any()).Return("", nil).MaxTimes(0),
					m.config.EXPECT().GetShortAddress().Return(defURL).MaxTimes(0),
				)
			},
			w: want{
				response: nil,
				err:      assert.Error,
				code:     codes.DataLoss,
			},
		},
		{
			name: "custom cutter error",
			request: &pb.CutterRequest{
				Url: "http://ya.ru",
			},
			pMocks: func(m *mock) {
				gomock.InOrder(
					m.cutter.EXPECT().Cut(gomock.Any(), gomock.Any()).Return("", errors.New("custom")).MaxTimes(1),
					m.config.EXPECT().GetShortAddress().Return(defURL).MaxTimes(0),
				)
			},
			w: want{
				response: nil,
				err:      assert.Error,
				code:     codes.Internal,
			},
		},

		{
			name: "UniqueURLError cutter",
			request: &pb.CutterRequest{
				Url: "http://ya.ru",
			},
			pMocks: func(m *mock) {
				gomock.InOrder(
					m.cutter.EXPECT().Cut(gomock.Any(), gomock.Any()).Return("", &errorUnique).MaxTimes(1),
					m.config.EXPECT().GetShortAddress().Return(defURL).MaxTimes(1),
				)
			},
			w: want{
				response: &pb.CutterResponse{
					Result: fmt.Sprintf(defURLPattern, defCode),
				},
				err:  assert.Error,
				code: codes.AlreadyExists,
			},
		},

		{
			name: "positive",
			request: &pb.CutterRequest{
				Url: "http://ya.ru",
			},
			pMocks: func(m *mock) {
				gomock.InOrder(
					m.cutter.EXPECT().Cut(gomock.Any(), gomock.Any()).Return(defCode, nil).MaxTimes(1),
					m.config.EXPECT().GetShortAddress().Return(defURL).MaxTimes(1),
				)
			},
			w: want{
				response: &pb.CutterResponse{
					Result: fmt.Sprintf(defURLPattern, defCode),
				},
				err: assert.NoError,
			},
		}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			a := mocks.NewMockICutter(ctrl)
			c := mocks.NewMockConfiger(ctrl)
			m := mock{
				config: c,
				cutter: a,
			}
			if tt.pMocks != nil {
				tt.pMocks(&m)
			}

			g := New(a, c)
			got, err := g.h.CutterJson(context.Background(), tt.request)

			if !tt.w.err(t, err, fmt.Sprintf("%s: %v", tt.name, tt.request)) {
				return
			}
			if err != nil {
				if e, ok := status.FromError(err); ok {
					assert.Equal(t, tt.w.code, e.Code())
				}
			}
			assert.Equalf(t, tt.w.response, got, "%s: %v", tt.name, tt.request)
		})
	}

}

//func TestCutterJson(t *testing.T) {
// 	type mock struct {
// 		config *mocks.MockConfiger
// 		cutter *mocks.MockICutter
// 	}

// 	type want struct {
// 		response *pb.CutterResponse
// 		err      assert.ErrorAssertionFunc
// 		code     codes.Code
// 	}

// 	tests := []struct {
// 		name    string
// 		request *pb.CutterRequest
// 		pMocks  func(m *mock)
// 		w       want
// 	}{{}}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			defer ctrl.Finish()
// 			a := mocks.NewMockICutter(ctrl)
// 			c := mocks.NewMockConfiger(ctrl)
// 			m := mock{
// 				config: c,
// 				cutter: a,
// 			}
// 			if tt.pMocks != nil {
// 				tt.pMocks(&m)
// 			}

// 			g := New(a, c)
// 			got, err := g.h.CutterJson(context.Background(), tt.request)

// 			if !tt.w.err(t, err, fmt.Sprintf("%s: %v", tt.name, tt.request)) {
// 				return
// 			}
// 			if err != nil {
// 				if e, ok := status.FromError(err); ok {
// 					assert.Equal(t, tt.w.code, e.Code())
// 				}
// 			}
// 			assert.Equalf(t, tt.w.response, got, "Short(_, %v)", tt.request)
// 		})
// 	}

// }
